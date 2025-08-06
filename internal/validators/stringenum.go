package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringEnumValidator struct {
	values []string
}

func (v stringEnumValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("field can have oneof %v", v.values)
}

func (v stringEnumValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("field can have oneof %v", v.values)
}

func (v stringEnumValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	value := req.ConfigValue.ValueString()
	for _, s := range v.values {
		if s == value {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid value",
		fmt.Sprintf("Field can oneof %v", v.values),
	)
}

func OneOf(values ...string) validator.String {
	return stringEnumValidator{values: values}
}
