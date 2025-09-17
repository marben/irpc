package main

// todo: remove the cyclic dependency to generator
// todo: probably best would be to contain all the import specs in here
type qualifier struct {
	g              *generator
	srcFileImports orderedSet[importSpec]
	usedImports    orderedSet[importSpec]
	tr             *typeResolver
}

func newQualifier(g *generator, tr *typeResolver, srcFileImports orderedSet[importSpec]) *qualifier {
	return &qualifier{
		g:              g,
		srcFileImports: srcFileImports,
		usedImports:    newOrderedSet[importSpec](),
		tr:             tr,
	}
}

func (q *qualifier) qualifyNamedInfo(ni namedInfo) string {
	// log.Printf("qualifying for importspec: %#v", ni.importSpec)
	qual := q.qualifierForImportSpec(ni.importSpec)
	if qual == "" {
		return ni.namedName
	}
	return qual + "." + ni.namedName
}

func (q *qualifier) qualifierForImportSpec(is importSpec) string {
	// log.Printf("qualifying for import spec %#v", is)
	// types with no import - int, float, string, ...
	if is == (importSpec{}) {
		return ""
	}

	// types from our own package
	if is.path == q.tr.inputPkg.PkgPath {
		return ""
	}

	// if pkg with exact alias is used in src, we use it's import and alias
	// given src imports were correct/non overlapping in the import file
	// we can just directly use them here
	for _, srcSpec := range q.srcFileImports.ordered {
		// if srcSpec.path == is.path {
		if srcSpec == is {
			q.addUsedImport(srcSpec)
			return srcSpec.packageQualifier()
		}
	}

	// ok, so at least the import path is used in the src imports?
	for _, srcSpec := range q.srcFileImports.ordered {
		if srcSpec.path == is.path {
			q.addUsedImport(srcSpec)
			return srcSpec.packageQualifier()
		}
	}

	// ok, so our import is from some dependency

	// if pkg is alredy marked as used (imported), we use given import
	for _, usedSpec := range q.usedImports.ordered {
		if usedSpec.path == is.path {
			return usedSpec.packageQualifier()
		}
	}

	// ok, new import.
	// need to be careful not to use already existing  alias/pkg name
	requestedQualifier := is.packageQualifier()
	for {
		if !q.qualifierUsedInSrcImports(requestedQualifier) && !q.qualifierUsedInUsedImports(requestedQualifier) {
			if requestedQualifier == is.pkgName {
				q.addUsedImport(is)
			} else {
				// we need to use alias
				spec := is
				spec.alias = requestedQualifier
				q.addUsedImport(spec)
			}

			return requestedQualifier
		}
		requestedQualifier += "_"
	}
}

/*
func (q *qualifier) qualifyType(t Type) string {
	// log.Printf("qualifying type: %#v", t)
	pkgQual := q.resolveTypePackageQualifier(t)
	if pkgQual == "" {
		// log.Printf("pkgQual == \"\". qualifying as: %q", t.Name())
		return t.Name()
	}

	return pkgQual + "." + t.Name()
}
*/

func (q *qualifier) addUsedImport(specs ...importSpec) {
	for _, spec := range specs {
		if spec != (importSpec{}) {
			q.usedImports.add(spec)
		}
	}
}

/*
// packageQualifier is the "time" in "time.Time"
// also handles aliases etc...
func (q *qualifier) resolveTypePackageQualifier(t Type) string {
	typeImportSpecs := t.ImportSpecs()
	// types with no import - int, float, string, ...
	if typeImportSpecs == (importSpec{}) {
		return ""
	}

	// types from our own package
	if typeImportSpecs.path == q.tr.inputPkg.PkgPath {
		return ""
	}

	// if pkg is used in the src file, we use it's import/alias
	// given src imports were correct/non overlapping in the import file
	// we can just directly use them here
	for _, srcSpec := range q.srcFileImports.ordered {
		if srcSpec.path == typeImportSpecs.path {
			q.addUsedImport(srcSpec)
			return srcSpec.packageQualifier()
		}
	}

	// ok, so our import is from some dependency

	// if pkg is alredy marked as used (imported), we use given import
	for _, usedSpec := range q.usedImports.ordered {
		if usedSpec.path == typeImportSpecs.path {
			return usedSpec.packageQualifier()
		}
	}

	// ok, new import.
	// need to be careful not to use already existing  alias/pkg name
	requestedQualifier := typeImportSpecs.packageQualifier()
	for {
		if !q.qualifierUsedInSrcImports(requestedQualifier) && !q.qualifierUsedInUsedImports(requestedQualifier) {
			if requestedQualifier == typeImportSpecs.pkgName {
				q.addUsedImport(typeImportSpecs)
			} else {
				// we need to use alias
				spec := typeImportSpecs
				spec.alias = requestedQualifier
				q.addUsedImport(spec)
			}

			return requestedQualifier
		}
		requestedQualifier += "_"
	}
}
*/

func (q *qualifier) qualifierUsedInUsedImports(qualifier string) bool {
	for _, srcImport := range q.srcFileImports.ordered {
		if srcImport.packageQualifier() == qualifier {
			return true
		}
	}
	return false
}

func (q *qualifier) qualifierUsedInSrcImports(qualifier string) bool {
	for _, usedImport := range q.usedImports.ordered {
		if usedImport.packageQualifier() == qualifier {
			return true
		}
	}
	return false
}
