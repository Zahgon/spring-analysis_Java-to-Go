// Command base is the transaction demonstration's entry point, ported from
// base.Boostrap.
//
// It asks config.xml's container for a TransactionBean. config.xml declares no
// component scan and no such bean, so the lookup fails — as it does in the
// original, which exits non-zero with the same message. The failure is the
// recorded behaviour and is not repaired here.
package main

import (
	"fmt"
	"os"

	"github.com/seaswalker/spring-analysis/internal/base/transaction"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"

	// config.xml names classes from the aop package; importing it registers
	// them, the way having them on the classpath did.
	_ "github.com/seaswalker/spring-analysis/internal/aop"
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
	bean, err := beans.GetBeanAs[*transaction.TransactionBean](ctx)
	if err != nil {
		return err
	}
	bean.Process()
	return nil
}
