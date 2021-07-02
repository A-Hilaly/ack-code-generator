// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Code generated by ack-generate. DO NOT EDIT.

package code

import (
	"fmt"
	"strings"

	awssdkmodel "github.com/aws/aws-sdk-go/private/model/api"

	"github.com/aws-controllers-k8s/code-generator/pkg/generate/multiversion"
	"github.com/aws-controllers-k8s/code-generator/pkg/model"
	"github.com/aws-controllers-k8s/code-generator/pkg/names"
)

var (
	builtinStatuses = []string{"ACKResourceMetadata", "Conditions"}
)

// ConvertTo return Go code that convert a spoke version CRD to
// a hub version.
func ConvertTo(
	src, dst *model.CRD,
	hubImportAlias string,
	srcVarName string,
	dstRawVarName string,
	indentLevel int,
) string {
	return convert(
		src, dst,
		true,
		hubImportAlias,
		srcVarName,
		dstRawVarName,
		indentLevel,
	)
}

// ConvertFrom return Go code that convert from a hub version CRD to
// a spoke Version.
func ConvertFrom(
	src, dst *model.CRD,
	hubImportAlias string,
	srcVarName string,
	dstRawVarName string,
	indentLevel int,
) string {
	return convert(
		src, dst,
		false,
		hubImportAlias,
		srcVarName,
		dstRawVarName,
		indentLevel,
	)
}

// convert outputs Go code that convert a source CRD into a destination crd.
func convert(
	src, dst *model.CRD,
	isCopyingToHub bool,
	hubImportAlias string,
	srcVarName string,
	dstRawVarName string,
	indentLevel int,
) string {
	out := "\n"
	indent := strings.Repeat("\t", indentLevel)

	copyFromVarName := "dst"
	copyFromRawVarName := "dstRaw"
	if !isCopyingToHub {
		copyFromVarName = "src"
		copyFromRawVarName = "srcRaw"
	}

	// cast srcRaw/dstRaw type
	// dst := dstRaw.(*v1.Repository) || src := srcRaw.(*v1.Repository)
	out += fmt.Sprintf(
		"%s%s := %s.(*%s.%s)\n",
		indent,
		copyFromVarName,
		copyFromRawVarName,
		hubImportAlias,
		dst.Names.Camel,
	)

	deltas, err := multiversion.ComputeCRDFieldsDeltas(src, dst)
	if err != nil {
		msg := fmt.Sprintf("delta computation error: %v", err)
		panic(msg)
	}

	// objectMetadataCopy := src.ObjectMeta
	out += fmt.Sprintf(
		"%sobjectMetadataCopy := src.ObjectMeta\n\n",
		indent,
	)

	toVarName := "dst.Spec"
	fromVarName := "src.Spec"
	out += generateFieldsDeltasCode(
		deltas.SpecDeltas,
		hubImportAlias,
		fromVarName,
		toVarName,
		isCopyingToHub,
		indentLevel,
	)
	toVarName = "dst.Status"
	fromVarName = "src.Status"
	out += generateFieldsDeltasCode(
		deltas.StatusDeltas,
		hubImportAlias,
		fromVarName,
		toVarName,
		isCopyingToHub,
		indentLevel,
	)

	for _, status := range builtinStatuses {
		// dst.Status.Conditions = src.Status.Conditions
		out += fmt.Sprintf("%s%s.%s = %s.%s\n", indent, toVarName, status, fromVarName, status)
	}

	out += "\n"
	// dst.ObjectMeta = objectMetadataCopy
	out += fmt.Sprintf("%sdst.ObjectMeta = objectMetadataCopy\n", indent)
	// return nil
	out += fmt.Sprintf("%sreturn nil", indent)

	return out
}

// generateFieldsDeltasCode translates FieldDeltas into Go code that
// converts a CRD to another.
func generateFieldsDeltasCode(
	deltas []multiversion.FieldDelta,
	hubImportAlias string,
	fromVarName string,
	toVarName string,
	isCopyingToHub bool,
	indentLevel int,
) string {
	out := ""
	for _, delta := range deltas {
		from := delta.Spoke
		to := delta.Hub

		switch delta.ChangeType {
		case multiversion.ChangeTypeIntact:
			out += copyField(
				from,
				to,
				hubImportAlias,
				fromVarName,
				toVarName,
				isCopyingToHub,
				indentLevel,
			)
		case multiversion.ChangeTypeRenamed:
			out += copyField(
				from,
				to,
				hubImportAlias,
				fromVarName,
				toVarName,
				isCopyingToHub,
				indentLevel,
			)
		case multiversion.ChangeTypeAdded:
			out += copyAddedRemovedField(
				to,
				hubImportAlias,
				fromVarName,
				toVarName,
				isCopyingToHub,
				indentLevel,
			)

		case multiversion.ChangeTypeRemoved:
			out += copyAddedRemovedField(
				from,
				hubImportAlias,
				fromVarName,
				toVarName,
				isCopyingToHub,
				indentLevel,
			)
		case multiversion.ChangeTypeShapeChanged, multiversion.ChangeTypeShapeChangedToSecret, multiversion.ChangeTypeUnknown:
			panic("Not implemented ChangeType in generate.code.generateFieldsDeltasCode")
		default:
			panic("Unsupported ChangeType in generate.code.generateFieldsDeltasCode")
		}
	}
	return out
}

// copyField outputs Go code that converts a CRD field that either stayed
// intact or was renamed.
//
// Output code will look something like this:
//
//   if src.Spec.ScanConfig != nil {
//       imageScanningConfigurationCopy := &v2.ImageScanningConfiguration{}
//       imageScanningConfigurationCopy.ScanOnPush = src.Spec.ScanConfig.ScanOnPush
//       dst.Spec.ScanConfig = imageScanningConfigurationCopy
//   }
func copyField(
	from, to *model.Field,
	hubImportAlias string,
	varFrom string,
	varTo string,
	isCopyingToHub bool,
	indentLevel int,
) string {
	// if a field is not renamed, from and to have the same name
	// so the name doesn't impact much the code generation
	varFromPath := varFrom + "." + from.Names.Camel
	varToPath := varTo + "." + to.Names.Camel

	// however in case of renames we should correctly invert from/to field
	// paths. Only when we are convert to hub (not from hub).
	if !isCopyingToHub {
		varFromPath = varFrom + "." + to.Names.Camel
		varToPath = varTo + "." + from.Names.Camel
	}

	switch from.ShapeRef.Shape.Type {
	case "structure":
		return copyStruct(
			from.ShapeRef.Shape,
			hubImportAlias,
			varFromPath,
			varToPath,
			isCopyingToHub,
			indentLevel,
		)
	case "list":
		return copyList(
			from.ShapeRef.Shape,
			hubImportAlias,
			varFromPath,
			varToPath,
			isCopyingToHub,
			indentLevel,
		)
	case "map":
		return copyMap(
			from.ShapeRef.Shape,
			hubImportAlias,
			varFromPath,
			varToPath,
			isCopyingToHub,
			indentLevel,
		)
	default:
		return copyScalar(
			from.ShapeRef.Shape,
			hubImportAlias,
			varFromPath,
			varToPath,
			indentLevel,
		)
	}
}

// copyAddedRemovedField outputs Go code that converts a CRD field that
// was either added or removed. The generated code uses on annotation
// to store a field data 'removed' field.
//
// Output code will look something like this:
//
//   annotationKey, annotationValueVar, err := AnnotateFieldData("EncryptionConfiguration", src.Spec.EncryptionConfiguration)
//   if err != nil {
// 	     return err
//   }
//   objectMetadataCopy.Annotations[annotationKey] = annotationValueVar
//
// Or
//
//   err := DecodeFieldDataAnnotation("conversions.ack.aws.dev/EncryptionConfiguration", dst.Spec.EncryptionConfiguration)
//   if err != nil {
// 	     return err
//   }
func copyAddedRemovedField(
	from *model.Field,
	hubImportAlias string,
	varFrom string,
	varTo string,
	isCopyingToHub bool,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)

	errVar := "err"
	annotationKeyVar := "annotationKey"
	annotationValueVar := "annotationValueVar"

	if !isCopyingToHub {
		// annotationKey, annotationValueVar, err := AnnotateFieldData("EncryptionConfiguration", src.Spec.EncryptionConfiguration)
		out += fmt.Sprintf(
			"%s%s, %s, %s := AnnotateFieldData(\"%s\", %s)\n",
			indent,
			annotationKeyVar,
			annotationValueVar,
			errVar,
			from.Names.Camel,
			varFrom+"."+from.Names.Camel,
		)
		// if err != nil { return err }
		out += fmt.Sprintf("%sif err != nil {\n", indent)
		out += fmt.Sprintf("%s\treturn err\n", indent)
		out += fmt.Sprintf("%s}\n", indent)
		// objectMetadataCopy.Annotations[annotationKey] = annotationValueVar
		out += fmt.Sprintf(
			"%sobjectMetadataCopy.Annotations[%s] = %s\n",
			indent,
			annotationKeyVar,
			annotationValueVar,
		)
	} else {
		// err := DecodeFieldDataAnnotation("conversions.ack.aws.dev/EncryptionConfiguration", dst.Spec.EncryptionConfiguration)
		out += fmt.Sprintf(
			"%s%s := DecodeFieldDataAnnotation(%s, %s)\n",
			indent,
			errVar,
			"\"conversions.ack.aws.dev/"+from.Names.Camel+"\"",
			varTo+"."+from.Names.Camel,
		)
		// if err != nil { return err }
		out += fmt.Sprintf("%sif err != nil {\n", indent)
		out += fmt.Sprintf("%s\treturn err\n", indent)
		out += fmt.Sprintf("%s}\n", indent)
	}
	return out
}

// copyScalar outputs Go code that converts a CRD field
// that is a scalar.
//
// Output code will look something like this:
//
//   dst.Status.CreatedAt = src.Status.CreatedAt
func copyScalar(
	shape *awssdkmodel.Shape,
	hubImportAlias string,
	varFrom string,
	varTo string,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)

	switch shape.Type {
	case "boolean", "string", "character", "byte", "short",
		"integer", "long", "float", "double", "timestamp":
		out += fmt.Sprintf(
			// dst.Spec.Conditions = src.Spec.Conditions
			"%s%s = %s\n",
			indent, varTo, varFrom,
		)
	default:
		panic("Unsupported shape type: " + shape.Type)
	}
	return out
}

// copyStruct outputs Go code that converts a struct to another.
//
// Output code will look something like this:
//
// elementCopy := &Tag{}
// if element != nil {
// 	   tagCopy := &Tag{}
// 	   tagCopy.Key = element.Key
// 	   tagCopy.Value = element.Value
// 	   elementCopy = tagCopy
// }
func copyStruct(
	shape *awssdkmodel.Shape,
	hubImportAlias string,
	varFrom string,
	varTo string,
	isCopyingToHub bool,
	//fieldPath string,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	//TODO(a-hilaly): use ackcompare.HasNilDifference
	// if src.Spec.Tags != nil {
	out += fmt.Sprintf(
		"%sif %s != nil {\n",
		indent,
		varFrom,
	)

	// initialize a new copy struct
	structShapeName := names.New(shape.ShapeName)
	varStructCopy := structShapeName.CamelLower + "Copy"
	out += newShapeTypeInstance(
		shape,
		hubImportAlias,
		varStructCopy,
		varFrom,
		true,
		isCopyingToHub,
		indentLevel+1,
	)

	// copy struct fields
	for _, memberName := range shape.MemberNames() {
		memberShapeRef := shape.MemberRefs[memberName]
		memberShape := memberShapeRef.Shape
		memberNames := names.New(memberName)

		switch memberShape.Type {
		case "structure":
			out += copyStruct(
				memberShape,
				hubImportAlias,
				varFrom+"."+memberNames.Camel,
				varStructCopy+"."+memberNames.Camel,
				isCopyingToHub,
				indentLevel+1,
			)
		case "list":
			out += copyList(
				memberShape,
				hubImportAlias,
				varFrom+"."+memberNames.Camel,
				varStructCopy+"."+memberNames.Camel,
				isCopyingToHub,
				indentLevel+1,
			)
		case "map":
			out += copyMap(
				memberShape,
				hubImportAlias,
				varFrom+"."+memberNames.Camel,
				varStructCopy+"."+memberNames.Camel,
				isCopyingToHub,
				indentLevel+1,
			)
		default:
			out += copyScalar(
				memberShape,
				hubImportAlias,
				varFrom+"."+memberNames.Camel,
				varStructCopy+"."+memberNames.Camel,
				indentLevel+1,
			)
		}
	}
	out += storeVariableIn(varStructCopy, varTo, indentLevel+1)
	out += fmt.Sprintf(
		"%s}\n", indent,
	)
	return out + "\n"
}

// copyList outputs Go code that copies one array to another.
//
// Output code will look something like this:
//
//   tagListCopy := make([]*v2.Tag, 0, len(src.Spec.Tags))
//   for i, element := range src.Spec.Tags {
//   	_ = i // non-used value guard.
//   	elementCopy := &v2.Tag{}
//   	if element != nil {
//   		tagCopy := &v2.Tag{}
//   		tagCopy.Key = element.Key
//   		tagCopy.Value = element.Value
//   		elementCopy = tagCopy
//   	}
//
//   	tagListCopy = append(tagListCopy, elementCopy)
//   }
//   dst.Spec.Tags = tagListCopy
func copyList(
	shape *awssdkmodel.Shape,
	hubImportAlias string,
	varFrom string,
	varTo string,
	isCopyingToHub bool,
	indentLevel int,
) string {
	indent := strings.Repeat("\t", indentLevel)
	if isMadeOfBuiltinTypes(shape) {
		return fmt.Sprintf(
			"%s%s = %s\n",
			indent,
			varFrom,
			varTo,
		)
	}

	out := ""
	//TODO(a-hilaly): use ackcompare.HasNilDifference
	out += fmt.Sprintf(
		"%sif %s != nil {\n",
		indent,
		varFrom,
	)

	// initialize a new copy struct
	structShapeName := names.New(shape.ShapeName)
	varStructCopy := structShapeName.CamelLower + "Copy"
	out += newShapeTypeInstance(
		shape,
		hubImportAlias,
		varStructCopy,
		varFrom,
		false,
		isCopyingToHub,
		indentLevel+1,
	)

	varIndex := "i"
	varElement := "element"
	out += fmt.Sprintf(
		"%s\tfor %s, %s := range %s {\n",
		indent,
		varIndex,
		varElement,
		varFrom,
	)

	out += fmt.Sprintf(
		"%s\t\t_ = %s // non-used value guard.\n",
		indent,
		varIndex,
	)

	memberShapeRef := shape.MemberRef
	memberShape := memberShapeRef.Shape

	varElementCopy := "element" + "Copy"
	out += newShapeTypeInstance(
		memberShape,
		hubImportAlias,
		varElementCopy,
		varElement,
		true,
		isCopyingToHub,
		indentLevel+2,
	)

	switch memberShape.Type {
	case "structure":
		out += copyStruct(
			memberShape,
			hubImportAlias,
			varElement,
			varElementCopy,
			isCopyingToHub,
			indentLevel+2,
		)
	case "list", "map":
		//TODO nested maps and maps of arrays
		if isMadeOfBuiltinTypes(memberShape) {

		}
	default:
		panic(fmt.Sprintf("Unsupported shape type in generate.code.copyMap"))
	}

	out += fmt.Sprintf(
		"%s\t\t%s = append(%s, %s)\n",
		indent,
		varStructCopy,
		varStructCopy,
		varElementCopy,
	)

	// closing loop
	out += fmt.Sprintf(
		"%s\t}\n", indent,
	)

	out += storeVariableIn(varStructCopy, varTo, indentLevel+1)

	// attach the copy struct to the dst variable
	out += fmt.Sprintf(
		"%s}\n", indent,
	)
	return out + "\n"
}

// copyMap outputs Go code that copies one map to another.
func copyMap(
	shape *awssdkmodel.Shape,
	hubImportAlias string,
	varFrom string,
	varTo string,
	isCopyingToHub bool,
	indentLevel int,
) string {
	indent := strings.Repeat("\t", indentLevel)
	if isMadeOfBuiltinTypes(shape) {
		return fmt.Sprintf(
			"%s%s = %s\n",
			indent,
			varFrom,
			varTo,
		)
	}

	out := ""
	out += fmt.Sprintf(
		"%sif %s != nil {\n",
		indent,
		varFrom,
	)

	structShapeName := names.New(shape.ShapeName)
	varStructCopy := structShapeName.CamelLower + "Copy"
	out += newShapeTypeInstance(
		shape,
		hubImportAlias,
		varStructCopy,
		varFrom,
		false,
		isCopyingToHub,
		indentLevel+1,
	)

	keyVarName := "k"
	valueVarName := "v"
	out += fmt.Sprintf(
		"%s\tfor %s, %s := range %s {\n",
		indent,
		keyVarName,
		valueVarName,
		varFrom,
	)

	out += fmt.Sprintf(
		"%s\t\t_ = %s // non-used value guard.\n",
		indent,
		keyVarName,
	)

	memberShapeRef := shape.MemberRef
	memberShape := memberShapeRef.Shape

	varElementCopy := "element" + "Copy"
	out += newShapeTypeInstance(
		memberShape,
		hubImportAlias,
		varElementCopy,
		valueVarName,
		true,
		isCopyingToHub,
		indentLevel+2,
	)

	switch memberShape.Type {
	case "structure":
		out += copyStruct(
			memberShape,
			hubImportAlias,
			valueVarName,
			varElementCopy,
			isCopyingToHub,
			indentLevel+2,
		)
	case "list", "map":
		//TODO nested maps and maps of arrays
		if isMadeOfBuiltinTypes(memberShape) {

		}
	default:
		panic(fmt.Sprintf("Unsupported shape type in generate.code.copyMap"))
	}

	out += fmt.Sprintf(
		"%s\t\t%s = map[(%s, %s)\n",
		indent,
		varStructCopy,
		varStructCopy,
		varElementCopy,
	)
	out += fmt.Sprintf(
		"%s\t}\n", indent,
	)

	out += storeVariableIn(varStructCopy, varTo, indentLevel+1)

	// attach the copy struct to the dst variable
	out += fmt.Sprintf(
		"%s}\n", indent,
	)

	return out + "\n"
}

// newShapeTypeInstance returns Go code that instanciate a new shape type.
//
// Output code will look something like this:
//
//   imageScanningConfigurationCopy := &v2.ImageScanningConfiguration{}
func newShapeTypeInstance(
	shape *awssdkmodel.Shape,
	hubImportAlias string,
	allocationVarName string,
	fromVar string,
	isPointer bool,
	isCopyingToHub bool,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)

	switch shape.Type {
	case "structure":
		goType := shape.GoTypeElem()
		if isCopyingToHub {
			goType = hubImportAlias + "." + goType
		}
		if isPointer {
			goType = "&" + goType
		}
		out += fmt.Sprintf(
			"%s%s := %s{}\n",
			indent,
			allocationVarName,
			goType,
		)
	case "list":
		goType := shape.MemberRef.GoTypeElem()
		if isCopyingToHub {
			goType = hubImportAlias + "." + goType
		}
		if isPointer {
			goType = "*" + goType
		}
		out += fmt.Sprintf(
			"%s%s := make([]*%s, 0, len(%s))\n",
			indent,
			allocationVarName,
			goType,
			fromVar,
		)
	case "map":
		goType := shape.ValueRef.GoTypeElem()
		if isCopyingToHub {
			goType = hubImportAlias + "." + goType
		}
		if isPointer {
			goType = "*" + goType
		}
		out += fmt.Sprintf(
			"%s%s := make(map[string]*%s, 0, len(%s))\n",
			indent,
			allocationVarName,
			goType,
			fromVar,
		)
	default:
		panic("Unsupported shape type in generate.code.newShapeTypeInstance")
	}

	return out
}

// TODO(remove)
func storeVariableIn(
	from string,
	target string,
	indentLevel int,
) string {
	out := ""
	indent := strings.Repeat("\t", indentLevel)
	out += fmt.Sprintf(
		"%s%s = %s\n",
		indent,
		target,
		from,
	)
	return out
}

// isMadeOfBuiltinTypes returns true if a given shape is fully made of Go
// builtin types.
func isMadeOfBuiltinTypes(shape *awssdkmodel.Shape) bool {
	switch shape.Type {
	case "boolean", "string", "character", "byte", "short",
		"integer", "long", "float", "double", "timestamp":
		return true
	case "list":
		return isMadeOfBuiltinTypes(shape.MemberRef.Shape)
	case "map":
		return isMadeOfBuiltinTypes(shape.ValueRef.Shape)
	default:
		return false
	}
}
