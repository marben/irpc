package main

type qualifier struct {
	srcFileImports orderedSet[importSpec]
	usedImports    orderedSet[importSpec]
	inputPkgPath   string
}

func newQualifier(tr typeResolver) *qualifier {
	return &qualifier{
		srcFileImports: tr.srcImports,
		usedImports:    newOrderedSet[importSpec](),
		inputPkgPath:   tr.inputPkg.PkgPath,
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
	if is.path == q.inputPkgPath {
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
		if !q.pkgQualifierUsedInSrcImports(requestedQualifier) && !q.pkgQualifierUsedInUsedImports(requestedQualifier) {
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

func (q *qualifier) addUsedImport(specs ...importSpec) {
	for _, spec := range specs {
		if spec != (importSpec{}) {
			q.usedImports.add(spec)
		}
	}
}

func (q *qualifier) pkgQualifierUsedInUsedImports(qualifier string) bool {
	for _, srcImport := range q.srcFileImports.ordered {
		if srcImport.packageQualifier() == qualifier {
			return true
		}
	}
	return false
}

// pkg qualifier is alias or actual package name
func (q *qualifier) pkgQualifierUsedInSrcImports(qualifier string) bool {
	for _, usedImport := range q.usedImports.ordered {
		if usedImport.packageQualifier() == qualifier {
			return true
		}
	}
	return false
}
