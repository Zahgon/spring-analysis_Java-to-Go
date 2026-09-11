// Command javaconfig is the annotation-configuration demonstration's entry
// point, ported from java_config.Bootrap.
//
// It builds a container from a configuration class, prints the name reached
// through the imported configuration, and then prints every bean definition
// the container registered — which is where the container's own naming and
// ordering become visible.
package main

import (
	"fmt"
	"os"

	"github.com/seaswalker/spring-analysis/internal/base"
	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	springctx "github.com/seaswalker/spring-analysis/internal/deps/springframework/context"
	"github.com/seaswalker/spring-analysis/internal/javaconfig"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, err := springctx.NewAnnotationConfigApplicationContext(javaconfig.NewSimpleBeanConfig())
	if err != nil {
		return err
	}
	simpleBean, err := beans.GetBeanAs[*base.SimpleBean](ctx)
	if err != nil {
		return err
	}
	fmt.Println(simpleBean.GetStudent().GetName())
	fmt.Println(javart.ArraysToString(ctx.GetBeanDefinitionNames()))
	return nil
}
