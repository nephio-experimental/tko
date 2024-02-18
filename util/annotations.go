package util

import (
	"github.com/tliron/go-ard"
)

const (
	MetadataAnnotation         = "nephio.org/metadata"
	MetadataAnnotationNever    = "Never"
	MetadataAnnotationHere     = "Here"
	MetadataAnnotationPostpone = "Postpone"

	MergeAnnotation         = "nephio.org/merge"
	MergeAnnotationReplace  = "Replace"
	MergeAnnotationOverride = "Override"

	RenameAnnotation = "nephio.org/rename"

	PrepareAnnotation         = "nephio.org/prepare"
	PrepareAnnotationNever    = "Never"
	PrepareAnnotationHere     = "Here"
	PrepareAnnotationPostpone = "Postpone"

	PreparedAnnotation = "nephio.org/prepared"
	ApprovedAnnotation = "nephio.org/approved"

	AnnotationTrue = "true"
)

func GetMetadataAnnotation(resource Resource) (string, bool) {
	return ard.With(resource).Get("metadata", "annotations", MetadataAnnotation).String()
}

func GetMergeAnnotation(resource Resource) (string, bool) {
	return ard.With(resource).Get("metadata", "annotations", MergeAnnotation).String()
}

func GetRenameAnnotation(resource Resource) (string, bool) {
	return ard.With(resource).Get("metadata", "annotations", RenameAnnotation).String()
}

func GetPrepareAnnotation(resource Resource) (string, bool) {
	return ard.With(resource).Get("metadata", "annotations", PrepareAnnotation).String()
}

func SetPrepareAnnotation(resource Resource, value string) bool {
	return ard.With(resource).ForceGet("metadata", "annotations", PrepareAnnotation).Set(value)
}

func GetPreparedAnnotation(resource Resource) (string, bool) {
	return ard.With(resource).Get("metadata", "annotations", PreparedAnnotation).String()
}

func IsPreparedAnnotation(resource Resource) bool {
	if prepared, ok := GetPreparedAnnotation(resource); ok {
		return prepared == AnnotationTrue
	}
	return false
}

func SetPreparedAnnotation(resource Resource, prepared bool) bool {
	annotation := ard.With(resource).ForceGet("metadata", "annotations", PreparedAnnotation)
	if prepared {
		if value, _ := annotation.String(); value == AnnotationTrue {
			return false
		}
		return annotation.Set(AnnotationTrue)
	} else {
		return annotation.Delete()
	}
}

func GetApprovedAnnotation(resource Resource) (string, bool) {
	return ard.With(resource).Get("metadata", "annotations", ApprovedAnnotation).String()
}

func IsApprovedAnnotation(resource Resource) bool {
	if approved, ok := GetApprovedAnnotation(resource); ok {
		return approved == AnnotationTrue
	}
	return false
}

func SetApprovedAnnotation(resource Resource, approved bool) bool {
	annotation := ard.With(resource).ForceGet("metadata", "annotations", ApprovedAnnotation)
	if approved {
		if value, _ := annotation.String(); value == AnnotationTrue {
			return false
		}
		return annotation.Set(AnnotationTrue)
	} else {
		return annotation.Delete()
	}
}

func UpdateAnnotationsForMerge(resource Resource) {
	annotation := ard.With(resource).Get("metadata", "annotations", MetadataAnnotation)
	if metadata, ok := annotation.String(); ok {
		switch metadata {
		case MetadataAnnotationHere, "":
			annotation.Set(MetadataAnnotationNever)
		case MetadataAnnotationPostpone:
			annotation.Set(MetadataAnnotationHere)
		case MetadataAnnotationNever:
			// Keep this annotation as is
		}
	}

	annotation = ard.With(resource).Get("metadata", "annotations", PrepareAnnotation)
	if prepare, ok := annotation.String(); ok {
		switch prepare {
		case PrepareAnnotationHere, "":
			// TODO: Is this necessary? We will never merge unprepared resources
			annotation.Set(PrepareAnnotationNever)
		case PrepareAnnotationPostpone:
			annotation.Set(PrepareAnnotationHere)
		case PrepareAnnotationNever:
			// Keep this annotation as is
		}
	}
}
