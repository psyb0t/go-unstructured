# Convert the 3.1 schema keywords used by both specs to OpenAPI 3.0 forms.
# Null union branches become nullable: true. Collapse singleton unions so
# oapi-codegen does not emit wrappers referring to undeclared branch types.
# const becomes a one-value enum, and schema examples arrays become example.
# UUID|string accepts every string. Flatten that redundant union because the
# generator makes path/query union parameters anonymous structs whose private
# storage cannot be populated by a caller.
# Preserve open-ended additional properties as raw JSON instead of decoding
# them into interface{}. This avoids implicit numeric conversion and the
# generated untyped deserialization flagged by Semgrep.
# Empty JSON response schemas also use raw JSON, preserving any valid JSON
# body without generating an interface{} response decoder.
walk(
  if type == "object" then
    (if .schema? == {} then
        .schema = {"x-go-type":"json.RawMessage"}
     else . end)
    | (if .additionalProperties? == true then
        .additionalProperties = {"x-go-type":"json.RawMessage"}
     else . end)
    | (if has("const") then .enum = [.const] | del(.const) else . end)
    | (if (.examples? | type) == "array" then
        .example = .examples[0] | del(.examples) else . end)
    | (if (.anyOf? | type) == "array" and
          any(.anyOf[]; .type == "null") then
        .anyOf |= map(select(.type != "null")) | .nullable = true
       else . end)
    | (if (.anyOf? | type) == "array" and
          all(.anyOf[]; .type == "string") and
          any(.anyOf[]; . == {"type":"string"}) then
        .type = "string" | del(.anyOf)
       else . end)
    | (if (.anyOf? | type) == "array" and (.anyOf | length) == 1 then
        .anyOf[0] + del(.anyOf) else . end)
  else . end
)
| .openapi = "3.0.3"
