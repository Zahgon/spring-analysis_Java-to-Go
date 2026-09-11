// Command jdkproxy is the dynamic-proxy demonstration's entry point, ported
// from test.proxy.JDKProxy.
//
// It wraps the service in a proxy and calls one method. The method the target
// calls on itself does not go back through the proxy, so the handler reports
// one call, not two.
package main

import "github.com/seaswalker/spring-analysis/internal/proxy"

func main() { run() }

func run() {
	userService := proxy.NewUserServiceImpl()
	proxied := proxy.NewUserServiceProxy(userService, proxy.NewHandler(userService))
	proxied.PrintName()
}
