// Command aop is the AOP demonstration's entry point, ported from
// aop.Bootstrap.
//
// It builds the container from config.xml, takes the advised bean out of it,
// calls the method that re-enters through its own proxy, and prints the
// runtime type it ended up holding.
package main

import (
	"fmt"
	"os"

	"github.com/seaswalker/spring-analysis/internal/aop"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, err := beans.NewClassPathXMLApplicationContext("config.xml")
	if err != nil {
		return err
	}
	bean, err := beans.GetBeanAs[aop.SimpleAopBeanAPI](ctx)
	if err != nil {
		return err
	}
	bean.TestB()
	fmt.Println(beans.SimpleClassName(bean))
	return nil
}
