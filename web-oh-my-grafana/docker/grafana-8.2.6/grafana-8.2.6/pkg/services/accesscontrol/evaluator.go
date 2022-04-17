package accesscontrol

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/grafana/grafana/pkg/infra/log"
)

var logger = log.New("accesscontrol.evaluator")

type Evaluator interface {
	// Evaluate permissions that are grouped by action
	Evaluate(permissions map[string]map[string]struct{}) (bool, error)
	// Inject params into the evaluator's templated scopes. e.g. "settings:" + eval.Parameters(":id") and returns a new Evaluator
	Inject(params map[string]string) (Evaluator, error)
	// String returns a string representation of permission required by the evaluator
	String() string
}

var _ Evaluator = new(permissionEvaluator)

// EvalPermission returns an evaluator that will require all scopes in combination with action to match
func EvalPermission(action string, scopes ...string) Evaluator {
	return permissionEvaluator{Action: action, Scopes: scopes}
}

type permissionEvaluator struct {
	Action string
	Scopes []string
}

func (p permissionEvaluator) Evaluate(permissions map[string]map[string]struct{}) (bool, error) {
	userScopes, ok := permissions[p.Action]
	if !ok {
		return false, nil
	}

	if len(p.Scopes) == 0 {
		return true, nil
	}

	for _, target := range p.Scopes {
		var err error
		var matches bool

		for scope := range userScopes {
			matches, err = match(scope, target)
			if err != nil {
				return false, err
			}
			if matches {
				break
			}
		}
		if !matches {
			return false, nil
		}
	}

	return true, nil
}

func match(scope, target string) (bool, error) {
	if scope == "" {
		return false, nil
	}

	if !ValidateScope(scope) {
		logger.Error(
			"invalid scope",
			"scope", scope,
			"reason", "scopes should not contain meta-characters like * or ?, except in the last position",
		)
		return false, nil
	}

	prefix, last := scope[:len(scope)-1], scope[len(scope)-1]
	//Prefix match
	if last == '*' {
		if strings.HasPrefix(target, prefix) {
			logger.Debug("matched scope", "user scope", scope, "target scope", target)
			return true, nil
		}
	}

	return scope == target, nil
}

func (p permissionEvaluator) Inject(params map[string]string) (Evaluator, error) {
	scopes := make([]string, 0, len(p.Scopes))
	for _, scope := range p.Scopes {
		tmpl, err := template.New("scope").Parse(scope)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err = tmpl.Execute(&buf, params); err != nil {
			return nil, err
		}
		scopes = append(scopes, buf.String())
	}
	return EvalPermission(p.Action, scopes...), nil
}

func (p permissionEvaluator) String() string {
	return fmt.Sprintf("action:%s scopes:%s", p.Action, strings.Join(p.Scopes, ", "))
}

var _ Evaluator = new(allEvaluator)

// EvalAll returns evaluator that requires all passed evaluators to evaluate to true
func EvalAll(allOf ...Evaluator) Evaluator {
	return allEvaluator{allOf: allOf}
}

type allEvaluator struct {
	allOf []Evaluator
}

func (a allEvaluator) Evaluate(permissions map[string]map[string]struct{}) (bool, error) {
	for _, e := range a.allOf {
		if ok, err := e.Evaluate(permissions); !ok || err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a allEvaluator) Inject(params map[string]string) (Evaluator, error) {
	var injected []Evaluator
	for _, e := range a.allOf {
		i, err := e.Inject(params)
		if err != nil {
			return nil, err
		}
		injected = append(injected, i)
	}
	return EvalAll(injected...), nil
}

func (a allEvaluator) String() string {
	permissions := make([]string, 0, len(a.allOf))
	for _, e := range a.allOf {
		permissions = append(permissions, e.String())
	}
	return fmt.Sprintf("all(%s)", strings.Join(permissions, " "))
}

var _ Evaluator = new(anyEvaluator)

// EvalAny returns evaluator that requires at least one of passed evaluators to evaluate to true
func EvalAny(anyOf ...Evaluator) Evaluator {
	return anyEvaluator{anyOf: anyOf}
}

type anyEvaluator struct {
	anyOf []Evaluator
}

func (a anyEvaluator) Evaluate(permissions map[string]map[string]struct{}) (bool, error) {
	for _, e := range a.anyOf {
		ok, err := e.Evaluate(permissions)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (a anyEvaluator) Inject(params map[string]string) (Evaluator, error) {
	var injected []Evaluator
	for _, e := range a.anyOf {
		i, err := e.Inject(params)
		if err != nil {
			return nil, err
		}
		injected = append(injected, i)
	}
	return EvalAny(injected...), nil
}

func (a anyEvaluator) String() string {
	permissions := make([]string, 0, len(a.anyOf))
	for _, e := range a.anyOf {
		permissions = append(permissions, e.String())
	}
	return fmt.Sprintf("any(%s)", strings.Join(permissions, " "))
}
