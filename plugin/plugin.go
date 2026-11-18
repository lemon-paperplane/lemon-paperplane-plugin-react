// Package react 是 Lemon Paperplane 的 React 插件：识别 React 与 Next.js / Create React App / Vite，
// 给出关键信息、默认脚本偏好，以及旧版 CRA / Next 在新 Node 上的预先提示与修复。
package react

import (
	"regexp"

	sdk "github.com/lemon-paperplane/lemon-paperplane-main/lemon-paperplane-main-plugin-sdk"
)

// PluginID 是插件 id。
const PluginID = "react"

// 构建工具。
const (
	BuildNext        = "next"
	BuildCRA         = "cra"
	BuildVite        = "vite"
	BuildReactRouter = "react-router"
)

// Plugin 实现 react 插件。
type Plugin struct{}

// New 返回插件实例。
func New() any { return &Plugin{} }

// Detection 是识别结果。
type Detection struct {
	React  bool
	Build  string
	Prefer []string
}

var (
	nextDevRe  = regexp.MustCompile(`\bnext(?:\s+dev\b|\s*$)`)
	craStartRe = regexp.MustCompile(`\b(?:react-scripts|react-app-rewired|craco)\s+start\b`)
	viteDevRe  = regexp.MustCompile(`\bvite\b(?:\s+(?:dev|serve)\b|\s*$|\s+--)`)
	rrDevRe    = regexp.MustCompile(`\b(?:react-router|remix(?:\s+vite:)?)\s+dev\b`)
)

// DetectReact 判断是否是 React 工程。
func DetectReact(pkg *pkgJSON) Detection {
	d := Detection{}
	if pkg == nil {
		return d
	}
	d.React = pkg.has("react") || pkg.has("next") || pkg.has("react-scripts")
	if !d.React {
		return d
	}
	switch {
	case pkg.has("next"):
		d.Build = BuildNext
		d.Prefer = scriptsMatching(pkg, nextDevRe.MatchString)
	case pkg.has("@react-router/dev") || pkg.has("@remix-run/dev"):
		d.Build = BuildReactRouter
		d.Prefer = scriptsMatching(pkg, rrDevRe.MatchString)
	case pkg.has("react-scripts"):
		d.Build = BuildCRA
		d.Prefer = scriptsMatching(pkg, craStartRe.MatchString)
	case pkg.has("vite"):
		d.Build = BuildVite
		d.Prefer = scriptsMatching(pkg, viteDevRe.MatchString)
	}
	return d
}

// Detect 在 nodejs facet 出现的目录里识别 React。
func (pl *Plugin) Detect(ctx *sdk.Context, p *sdk.DetectParams) (*sdk.DetectResult, error) {
	if !sdk.HasFile(p.Files, "package.json") {
		return &sdk.DetectResult{}, nil
	}
	pkg, err := readPkg(p.Dir)
	if err != nil {
		return &sdk.DetectResult{}, nil
	}
	d := DetectReact(pkg)
	if !d.React {
		return &sdk.DetectResult{}, nil
	}
	data := map[string]any{"build": d.Build}
	if len(d.Prefer) > 0 {
		data[sdk.FacetDataPreferScripts] = d.Prefer
	}
	res := &sdk.DetectResult{Facets: []sdk.Facet{{ID: "react", Data: data}}}
	if d.Build == BuildNext {
		res.Facets = append(res.Facets, sdk.Facet{ID: "nextjs"})
	}
	return res, nil
}

// Analyze 给出 React / 构建工具版本；react-scripts < 5、Next < 12 在 Node 17+ 上预先提示。
func (pl *Plugin) Analyze(ctx *sdk.Context, p *sdk.AnalyzeParams) (*sdk.AnalyzeResult, error) {
	dir := p.Module.Dir
	pkg, err := readPkg(dir)
	if err != nil {
		return &sdk.AnalyzeResult{}, nil
	}
	d := DetectReact(pkg)
	if !d.React {
		return &sdk.AnalyzeResult{}, nil
	}
	nm := loadNodeModel(&p.Module)
	res := &sdk.AnalyzeResult{}
	if v := depVersion(dir, pkg, nm, "react"); v != "" {
		res.Info = append(res.Info, sdk.InfoItem{Label: "React", Value: majorMinor(v), Icon: "icons/react.svg"})
	}
	var buildPkg, buildLabel, buildIcon string
	switch d.Build {
	case BuildNext:
		buildPkg, buildLabel, buildIcon = "next", "Next.js", "icons/nextjs.svg"
	case BuildCRA:
		buildPkg, buildLabel, buildIcon = "react-scripts", "Create React App", "icons/react.svg"
	case BuildReactRouter:
		buildPkg, buildLabel, buildIcon = "@react-router/dev", "React Router", "icons/react.svg"
		if !pkg.has(buildPkg) {
			buildPkg, buildLabel = "@remix-run/dev", "Remix"
		}
	case BuildVite:
		buildPkg, buildLabel, buildIcon = "vite", "Vite", "icons/vite.svg"
	}
	buildVer := ""
	if buildPkg != "" {
		buildVer = depVersion(dir, pkg, nm, buildPkg)
		res.Info = append(res.Info, sdk.InfoItem{Label: buildLabel, Value: majorMinor(buildVer), Icon: buildIcon})
	}
	if nm.LegacyPackage == "react-scripts" || nm.LegacyPackage == "next" {
		if nv, ok := legacyNeeded(ctx, p, nm); ok {
			what := buildLabel
			if mm := majorMinor(buildVer); mm != "" {
				what += " " + mm
			}
			res.Problems = append(res.Problems, legacyProblem(ctx, what, nv))
		}
	}
	return res, nil
}
