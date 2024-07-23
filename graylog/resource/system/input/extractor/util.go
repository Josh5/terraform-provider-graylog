package extractor

import (
	"encoding/json"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/josh5/terraform-provider-graylog/graylog/convert"
	"github.com/josh5/terraform-provider-graylog/graylog/util"
)

const (
	keyInputID         = "input_id"
	keyExtractorID     = "extractor_id"
	keyExtractorConfig = "extractor_config"
	keyID              = "id"
	keyConfig          = "config"
	keyConverters      = "converters"
	keyType            = "type"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func getDataFromResourceData(d *schema.ResourceData) (map[string]interface{}, error) {
	log.Println("[DEBUG] Entering getDataFromResourceData")

	data, err := convert.GetFromResourceData(d, Resource())
	if err != nil {
		log.Printf("[ERROR] Error converting resource data: %s", err)
		return nil, err
	}
	log.Printf("[DEBUG] Data after conversion: %+v", data)
	util.RenameKey(data, "type", "extractor_type")
	util.SetDefaultValue(data, "target_field", "")
	util.SetDefaultValue(data, "condition_value", "")

	if err := convert.JSONToData(data, keyExtractorConfig); err != nil {
		log.Printf("[ERROR] Error converting JSON to data: %s", err)
		return nil, err
	}
	util.RenameKey(data, keyExtractorID, keyID)

	log.Printf("[DEBUG] Data before handling converters: %+v", data)
	if conv, ok := data[keyConverters].([]interface{}); ok {
		for i, c := range conv {
			converter := c.(map[string]interface{})
			if err := convert.JSONToData(converter, keyConfig); err != nil {
				log.Printf("[ERROR] Error converting JSON to data for converters: %s", err)
				return nil, err
			}
			conv[i] = converter
		}
		data[keyConverters] = conv
	}

	log.Printf("[DEBUG] Final data: %+v", data)
	return data, nil
}

func setDataToResourceData(d *schema.ResourceData, data map[string]interface{}) error {
	log.Println("[DEBUG] Entering setDataToResourceData")
	log.Printf("[DEBUG] Data before converting to JSON: %+v", data)

	if err := convert.DataToJSON(data, keyExtractorConfig); err != nil {
		log.Printf("[ERROR] Error converting data to JSON: %s", err)
		return err
	}
	util.RenameKey(data, keyID, keyExtractorID)

	log.Printf("[DEBUG] Data after converting to JSON: %+v", data)
	if conv, ok := data[keyConverters].([]interface{}); ok {
		for i, c := range conv {
			converter := c.(map[string]interface{})
			if b, err := json.Marshal(converter[keyConfig]); err != nil {
				log.Printf("[ERROR] Error marshaling JSON for converters: %s", err)
				return err
			} else {
				converter[keyConfig] = string(b)
			}
			conv[i] = converter
		}
		data[keyConverters] = conv
	}

	log.Printf("[DEBUG] Final data before setting resource: %+v", data)

	if err := convert.SetResourceData(d, Resource(), data); err != nil {
		log.Printf("[ERROR] Error setting resource data: %s", err)
		return err
	}

	d.SetId(d.Get(keyInputID).(string) + "/" + d.Get(keyExtractorID).(string))
	log.Printf("[DEBUG] Resource ID set to: %s", d.Id())
	return nil
}
