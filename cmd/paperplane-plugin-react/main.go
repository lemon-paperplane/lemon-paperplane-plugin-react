// paperplane-plugin-react 是 react 插件的独立可执行入口。
package main

import (
	sdk "github.com/lemon-paperplane/lemon-paperplane-main/lemon-paperplane-main-plugin-sdk"
	react "github.com/lemon-paperplane/lemon-paperplane-plugin-react/plugin"
)

func main() { sdk.Serve(react.New()) }
