package main

type Scope struct {
	hashmap      map[string]*Variable
	lastDecl     *Variable
	currentDepth int
}

type Variable struct {
	name     string
	label    string
	kind     Type
	prevName *Variable
	prevDecl *Variable
	depth    int
}

func InitScope() Scope {
	return Scope{hashmap: make(map[string]*Variable)}
}

func (scope *Scope) declare(name string, kind Type) string {
	prev, prs := scope.hashmap[name]
	variable := Variable{}
	if prs {
		variable = Variable{name, prev.label + "_", kind, prev, scope.lastDecl, scope.currentDepth}
	} else {
		variable = Variable{name, name, kind, nil, scope.lastDecl, scope.currentDepth}
	}
	scope.hashmap[name] = &variable
	scope.lastDecl = &variable
	return variable.label
}

func (scope *Scope) get(name string) (string, Type, bool) {
	variable, prs := scope.hashmap[name]
	if prs {
		return variable.label, variable.kind, true
	}
	return "", 0, false
}

func (scope *Scope) pushScope() {
	scope.currentDepth++
}

func (scope *Scope) popScope() {
	scope.currentDepth--
	for scope.lastDecl != nil && scope.lastDecl.depth > scope.currentDepth {
		variable := scope.hashmap[scope.lastDecl.name]
		if variable.prevName != nil {
			scope.hashmap[scope.lastDecl.name] = variable.prevName
		} else {
			delete(scope.hashmap, scope.lastDecl.name)
		}
		scope.lastDecl = scope.lastDecl.prevDecl
	}
}
