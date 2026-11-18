package react

import (
	"encoding/json"
	"strings"
	"testing"

	sdk "github.com/lemon-paperplane/lemon-paperplane-main/lemon-paperplane-main-plugin-sdk"
)

const nextApp = `{"name":"web","scripts":{"dev":"next dev --turbopack","build":"next build","start":"next start"},"dependencies":{"next":"15.1.0","react":"^19.0.0","react-dom":"^19.0.0"}}`
const craOld = `{"name":"cra","scripts":{"start":"react-scripts start","build":"react-scripts build","test":"react-scripts test"},"dependencies":{"react":"^17.0.2","react-scripts":"4.0.3"}}`
const viteReact = `{"name":"vr","scripts":{"dev":"vite","build":"tsc -b && vite build","preview":"vite preview"},"dependencies":{"react":"^19.1.0"},"devDependencies":{"vite":"^6.3.5","@vitejs/plugin-react":"^4.4.1"}}`

func detect(t *testing.T, pkg string) *sdk.DetectResult {
	t.Helper()
	dir := writeFiles(t, map[string]string{"package.json": pkg})
	r, err := (&Plugin{}).Detect(testCtx(), &sdk.DetectParams{Dir: dir, Files: []string{"package.json"}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestDetect(t *testing.T) {
	r := detect(t, nextApp)
	if len(r.Facets) != 2 || r.Facets[1].ID != "nextjs" || strings.Join(r.Facets[0].Data["preferScripts"].([]string), ",") != "dev" {
		t.Errorf("next: %+v", r)
	}
	r = detect(t, craOld)
	if r.Facets[0].Data["build"] != BuildCRA || strings.Join(r.Facets[0].Data["preferScripts"].([]string), ",") != "start" {
		t.Errorf("cra: %+v", r)
	}
	r = detect(t, viteReact)
	if r.Facets[0].Data["build"] != BuildVite || strings.Join(r.Facets[0].Data["preferScripts"].([]string), ",") != "dev" {
		t.Errorf("vite: %+v", r)
	}
	r = detect(t, `{"scripts":{"build":"react-router build","dev":"react-router dev"},"dependencies":{"react":"^19.1.0","react-router":"^7.5.3"},"devDependencies":{"@react-router/dev":"^7.5.3","vite":"^6.3.3"}}`)
	if r.Facets[0].Data["build"] != BuildReactRouter || strings.Join(r.Facets[0].Data["preferScripts"].([]string), ",") != "dev" {
		t.Errorf("react router: %+v", r)
	}
	if r = detect(t, `{"dependencies":{"vue":"^3"}}`); len(r.Facets) != 0 {
		t.Errorf("vue 工程不是 react：%+v", r)
	}
}

func analyze(t *testing.T, pkg, nodeVer string, model json.RawMessage) *sdk.AnalyzeResult {
	t.Helper()
	dir := writeFiles(t, map[string]string{"package.json": pkg})
	p := &sdk.AnalyzeParams{Module: sdk.ModuleInfo{Dir: dir, Models: map[string]json.RawMessage{"nodejs": model}}, Toolchains: map[string]sdk.Toolchain{"node": {Kind: "node", Version: nodeVer}}}
	r, err := (&Plugin{}).Analyze(testCtx(), p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestAnalyze(t *testing.T) {
	r := analyze(t, nextApp, "22.20.0", nodeModelJSON(map[string]string{"next": "15.1.0", "react": "19.0.0"}, "^18.18 || ^19.8 || >=20", "26", false, ""))
	if len(r.Info) != 2 || r.Info[0].Value != "19.0" || r.Info[1].Label != "Next.js" || r.Info[1].Value != "15.1" || len(r.Problems) != 0 {
		t.Errorf("next: %+v %+v", r.Info, r.Problems)
	}
	m := nodeModelJSON(map[string]string{"react-scripts": "4.0.3", "react": "17.0.2"}, "", "16", true, "react-scripts")
	r = analyze(t, craOld, "22.20.0", m)
	if len(r.Problems) != 1 || r.Problems[0].RuleID != "openssl-legacy.static" || r.Problems[0].Fixes[0].Params["key"] != "NODE_OPTIONS" {
		t.Errorf("cra 4 / node 22: %+v", r.Problems)
	}
	dir := writeFiles(t, map[string]string{"package.json": craOld})
	pr := &sdk.AnalyzeParams{Module: sdk.ModuleInfo{Dir: dir, Models: map[string]json.RawMessage{"nodejs": m}},
		Toolchains: map[string]sdk.Toolchain{"node": {Kind: "node", Version: "22.20.0"}}, Env: map[string]string{"NODE_OPTIONS": "--openssl-legacy-provider"}}
	if r, _ := (&Plugin{}).Analyze(testCtx(), pr); len(r.Problems) != 0 {
		t.Errorf("已开启旧版兼容仍提示：%+v", r.Problems)
	}
	if r = analyze(t, craOld, "16.17.1", m); len(r.Problems) != 0 {
		t.Errorf("node 16: %+v", r.Problems)
	}
}

func TestRules(t *testing.T) {
	if ok, v := ruleMatch(t, "react.cra-port", "Something is already running on port 3000.", "port"); !ok || v != "3000" {
		t.Errorf("cra port %v %q", ok, v)
	}
	line := `You are using Node.js 18.17.0. For Next.js, Node.js version "^18.18.0 || ^19.8.0 || >= 20.0.0" is required.`
	if ok, v := ruleMatch(t, "react.next-node-range", line, "constraint"); !ok || v != "^18.18.0 || ^19.8.0 || >= 20.0.0" || !sdk.MatchVersion(v, "20.1.0") {
		t.Errorf("next range %v %q", ok, v)
	}
	if ok, v := ruleMatch(t, "react.next-node-min", "You are using Node.js 16.13.0. For Next.js, Node.js version >= v18.17.0 is required.", "constraint"); !ok || v != ">=18.17.0" {
		t.Errorf("next min %v %q", ok, v)
	}
}
