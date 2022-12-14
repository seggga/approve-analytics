// Package public GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package public

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/delays": {
            "get": {
                "security": [
                    {
                        "Auth": []
                    }
                ],
                "description": "Get delays on all finished and declined tasks",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get delays",
                "operationId": "delays",
                "responses": {
                    "200": {
                        "description": "task id and lag",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Delay"
                            }
                        }
                    },
                    "500": {
                        "description": "internal error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/totals": {
            "get": {
                "security": [
                    {
                        "Auth": []
                    }
                ],
                "description": "Get total amount of finished and declined tasks",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "analytics"
                ],
                "summary": "Get total counts",
                "operationId": "totals",
                "responses": {
                    "200": {
                        "description": "finished and declined task counters",
                        "schema": {
                            "$ref": "#/definitions/models.Totals"
                        }
                    },
                    "500": {
                        "description": "internal error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.Delay": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "lag": {
                    "type": "integer"
                }
            }
        },
        "models.Totals": {
            "type": "object",
            "properties": {
                "declined": {
                    "type": "integer"
                },
                "finished": {
                    "type": "integer"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Analytics service",
	Description:      "Signed token protects our admin endpoints",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
