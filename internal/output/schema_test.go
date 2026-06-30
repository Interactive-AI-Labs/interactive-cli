package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintSchemaPretty(t *testing.T) {
	raw := json.RawMessage(`{
		"title": "AgentConfig",
		"description": "Configuration for the Interactive Agent.",
		"type": "object",
		"required": ["context"],
		"properties": {
			"llm_model": {
				"anyOf": [
					{"minLength": 1, "type": "string"},
					{"type": "null"}
				],
				"default": null,
				"description": "Model identifier in provider/model form."
			},
			"context": {
				"$ref": "#/$defs/AgentContext"
			},
			"mcps": {
				"items": {"$ref": "#/$defs/McpConfig"},
				"type": "array"
			},
			"search": {
				"anyOf": [
					{
						"oneOf": [
							{"$ref": "#/$defs/KnowledgeBase"},
							{"$ref": "#/$defs/ExternalSearch"}
						],
						"discriminator": {
							"propertyName": "type",
							"mapping": {
								"knowledge_base": "#/$defs/KnowledgeBase",
								"external_search": "#/$defs/ExternalSearch"
							}
						}
					},
					{"type": "null"}
				],
				"default": null,
				"description": "Retrieval-grounding source."
			}
		},
		"$defs": {
			"AgentContext": {
				"title": "AgentContext",
				"description": "Behavioral configuration.",
				"type": "object",
				"required": ["description", "language"],
				"properties": {
					"description": {
						"$ref": "#/$defs/PromptRef"
					},
					"language": {
						"description": "Language the agent communicates in.",
						"type": "string"
					}
				}
			},
			"McpConfig": {
				"title": "McpConfig",
				"description": "MCP server connection details.",
				"type": "object",
				"required": ["id", "hostname", "port", "transport"],
				"properties": {
					"id": {
						"description": "Unique identifier.",
						"type": "string"
					},
					"hostname": {
						"description": "MCP server hostname.",
						"type": "string"
					},
					"port": {
						"type": "integer",
						"minimum": 1,
						"maximum": 65535
					},
					"transport": {
						"enum": ["sse", "streamable-http"],
						"type": "string"
					}
				}
			},
			"PromptRef": {
				"title": "PromptRef",
				"description": "A reference to a prompt at a specific version.",
				"type": "object",
				"required": ["id", "version"],
				"properties": {
					"id": {
						"description": "Prompt identifier or slug.",
						"type": "string"
					},
					"version": {
						"anyOf": [
							{"type": "integer"},
							{"type": "string"}
						],
						"description": "Version to bind to."
					}
				}
			}
		}
	}`)

	var buf bytes.Buffer
	err := PrintSchemaPretty(&buf, raw, "2.1.0")
	if err != nil {
		t.Fatalf("PrintSchemaPretty() error = %v", err)
	}

	out := buf.String()

	checks := []string{
		"Schema version: 2.1.0",
		"AgentConfig",
		"context",
		"AgentContext",
		"yes",
		"no",
		"list[McpConfig]",
		"string | null",
		"KnowledgeBase | ExternalSearch | null",
		"McpConfig",
		"PromptRef",
		"FIELD",
		"TYPE",
		"REQUIRED",
		"DEFAULT",
		"DESCRIPTION",
	}

	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("output missing %q\n\nfull output:\n%s", c, out)
		}
	}
}
