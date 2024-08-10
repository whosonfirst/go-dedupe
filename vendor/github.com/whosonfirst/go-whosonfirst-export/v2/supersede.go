package export

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func SupersedeRecord(ctx context.Context, ex Exporter, old_body []byte) ([]byte, []byte, error) {

	id_rsp := gjson.GetBytes(old_body, "properties.wof:id")

	if !id_rsp.Exists() {
		return nil, nil, fmt.Errorf("failed to derive old properties.wof:id property for record being superseded")
	}

	old_id := id_rsp.Int()

	// Create the new record

	new_body := old_body

	new_body, err := sjson.DeleteBytes(new_body, "properties.wof:id")

	if err != nil {
		return nil, nil, err
	}

	new_body, err = ex.Export(ctx, new_body)

	if err != nil {
		return nil, nil, err
	}

	id_rsp = gjson.GetBytes(new_body, "properties.wof:id")

	if !id_rsp.Exists() {
		return nil, nil, fmt.Errorf("failed to derive new properties.wof:id property for record superseding '%d'", old_id)
	}

	new_id := id_rsp.Int()

	// Update the new record

	new_body, err = sjson.SetBytes(new_body, "properties.wof:supersedes", []int64{old_id})

	if err != nil {
		return nil, nil, err
	}

	// Update the old record

	to_update := map[string]interface{}{
		"properties.mz:is_current":     0,
		"properties.wof:superseded_by": []int64{new_id},
	}

	old_body, err = AssignProperties(ctx, old_body, to_update)

	if err != nil {
		return nil, nil, err
	}

	return old_body, new_body, nil
}

func SupersedeRecordWithParent(ctx context.Context, ex Exporter, to_supersede_f []byte, parent_f []byte) ([]byte, []byte, error) {

	id_rsp := gjson.GetBytes(parent_f, "properties.wof:id")

	if !id_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent feature is missing properties.wof:id")
	}

	parent_id := id_rsp.Int()

	hier_rsp := gjson.GetBytes(parent_f, "properties.wof:hierarchy")

	if !hier_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent feature is missing properties.wof:hierarchy")
	}

	parent_hierarchy := hier_rsp.Value()

	inception_rsp := gjson.GetBytes(parent_f, "properties.edtf:inception")

	if !inception_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent record is missing properties.edtf:inception")
	}

	cessation_rsp := gjson.GetBytes(parent_f, "properties.edtf:cessation")

	if !cessation_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent record is missing properties.edtf:cessation")
	}

	inception := inception_rsp.String()
	cessation := cessation_rsp.String()

	to_update_old := map[string]interface{}{
		"properties.edtf:cessation": inception,
	}

	to_update_new := map[string]interface{}{
		"properties.wof:parent_id":  parent_id,
		"properties.wof:hierarchy":  parent_hierarchy,
		"properties.edtf:inception": inception,
		"properties.edtf:cessation": cessation,
	}

	//

	superseded_f, superseding_f, err := SupersedeRecord(ctx, ex, to_supersede_f)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to supersede record: %v", err)
	}

	superseded_f, err = AssignProperties(ctx, superseded_f, to_update_old)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to assign properties for new record: %v", err)
	}

	name_rsp := gjson.GetBytes(superseding_f, "properties.wof:name")

	if !name_rsp.Exists() {
		return nil, nil, fmt.Errorf("failed to retrieve wof:name for new record")
	}

	name := name_rsp.String()
	label := fmt.Sprintf("%s (%s)", name, inception)

	to_update_new["properties.wof:label"] = label

	superseding_f, err = AssignProperties(ctx, superseding_f, to_update_new)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to assign updated properties for new record %v", err)
	}

	return superseded_f, superseding_f, nil

}
