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

	PreparedAnnotation     = "nephio.org/prepared"
	PreparedAnnotationTrue = "true"
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
		return prepared == PreparedAnnotationTrue
	}
	return false
}

func SetPreparedAnnotation(resource Resource, prepared bool) bool {
	annotation := ard.With(resource).ForceGet("metadata", "annotations", PreparedAnnotation)
	if prepared {
		return annotation.Set(PreparedAnnotationTrue)
	} else {
		return annotation.Delete()
		// delete(annotations, PrepareAnnotation) // TODO: should we do this?
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
