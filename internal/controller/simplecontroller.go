// Package controller holds the web application's request handlers.
package controller

import (
	"fmt"

	"github.com/seaswalker/spring-analysis/internal/deps/javart"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/validation"
	"github.com/seaswalker/spring-analysis/internal/deps/springframework/web"
	"github.com/seaswalker/spring-analysis/internal/model"
)

// ViewName is the logical view both handlers render.
const ViewName = "echo"

// SimpleController is the @Controller the component scan picks up.
type SimpleController struct {
	validator *validation.JSR380Validator
}

// NewSimpleController returns the controller.
func NewSimpleController() *SimpleController { return &SimpleController{} }

// InitBinder is the @InitBinder hook.
//
// It is empty in the original — both lines that would register a validator are
// commented out — so the binder is left with none, and SimpleModelValidator
// never runs in the web flow. The hook is kept because its presence, and its
// doing nothing, is what the original demonstrates. The binder is named and
// deliberately left untouched.
func (c *SimpleController) InitBinder(binder *web.DataBinder) {
	_ = binder
}

// PostConstruct builds the JSR-380 validator, as the original's
// @PostConstruct initValidator does.
func (c *SimpleController) PostConstruct() error {
	c.validator = validation.BuildDefaultValidatorFactory()
	return nil
}

// Echo handles /echo: it greets the required name parameter.
func (c *SimpleController) Echo(req *web.Request) (string, error) {
	name, err := req.RequestParam("name")
	if err != nil {
		return "", err
	}
	req.Model.AddAttribute("echo", "hello "+name)
	return ViewName, nil
}

// EchoAgain handles POST /echoAgain: it binds the JSON body, reports every
// constraint violation on stdout, and greets the model.
//
// The violations are printed and then discarded. They are not added to the
// binding result, and the bound argument is not @Valid, so an age over the
// declared maximum still produces the success message — the "Ops, error!"
// branch is unreachable through this route. That asymmetry is the original's,
// and repairing it here would change what the demonstration shows.
func (c *SimpleController) EchoAgain(req *web.Request) (string, error) {
	simpleModel := model.NewSimpleModel()
	if err := req.RequestBody(simpleModel); err != nil {
		return "", err
	}
	bindingResult := validation.NewBindingResult(simpleModel, "simpleModel")

	for _, violation := range c.validator.Validate(simpleModel) {
		fmt.Println("错误消息: " + violation.GetMessage())
	}

	var hello string
	if bindingResult.HasErrors() {
		hello = "Ops, error!"
	} else {
		hello = fmt.Sprintf("hello %s, your age is %s.",
			simpleModel.GetName(), javart.ToString(simpleModel.GetAge()))
	}
	req.Model.AddAttribute("echo", hello)
	req.BindingResult = bindingResult
	fmt.Println(simpleModel)
	return ViewName, nil
}

// RequestMappings declares the routes the original's @RequestMapping
// annotations carry: /echo on any verb, /echoAgain on POST only.
func (c *SimpleController) RequestMappings() []web.HandlerMapping {
	return []web.HandlerMapping{
		{Mapping: web.RequestMapping{Path: "/echo"}, Handler: c.Echo},
		{Mapping: web.RequestMapping{Path: "/echoAgain", Methods: []string{"POST"}}, Handler: c.EchoAgain},
	}
}

var (
	_ web.Controller    = (*SimpleController)(nil)
	_ web.PostConstruct = (*SimpleController)(nil)
	_ web.InitBinder    = (*SimpleController)(nil)
)
