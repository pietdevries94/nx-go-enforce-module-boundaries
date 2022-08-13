package analyzer

const (
	messageProjectWithoutTagsCannotHaveDependencies = `A project without tags matching at least one constraint cannot depend on any libraries`
	messageNotTagsConstraintViolation               = `A project tagged with "%s" can not depend on libs tagged with %s`
	messageOnlyTagsConstraintViolation              = `A project tagged with "%s" can only depend on libs tagged with %s`
)
