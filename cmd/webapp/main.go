// Command webapp serves the Spring MVC demonstration, replacing the WAR the
// original deployed into Tomcat.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	"github.com/seaswalker/spring-analysis/internal/webapp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	port := flag.Int("port", webapp.DefaultPort, "port to listen on")
	viewRoot := flag.String("views", "web", "directory holding the view templates and static resources")
	classPath := flag.String("resources", beans.ClassPath, "classpath root holding the Spring configuration")
	flag.Parse()

	beans.ClassPath = *classPath

	dispatcher, err := webapp.New(*viewRoot)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Serving %s at http://localhost%s%s/\n", dispatcher.ContextPath(), addr, dispatcher.ContextPath())
	if err := http.ListenAndServe(addr, dispatcher); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
