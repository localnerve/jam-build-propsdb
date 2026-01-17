// app_data.go
//
// A scalable, high performance drop-in replacement for the jam-build nodejs data service
// Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
//
// This file is part of jam-build-propsdb.
// jam-build-propsdb is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later version.
// jam-build-propsdb is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
// You should have received a copy of the GNU Affero General Public License along with jam-build-propsdb.
// If not, see <https://www.gnu.org/licenses/>.
// Additional terms under GNU AGPL version 3 section 7:
// a) The reasonable legal notice of original copyright and author attribution must be preserved
//    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
//    in this material, copies, or source code of derived works.

package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/jam-build-propsdb/internal/services"
	"github.com/localnerve/jam-build-propsdb/internal/types"
	"github.com/localnerve/jam-build-propsdb/internal/utils"
	"gorm.io/gorm"
)

// AppDataHandler handles application data routes
type AppDataHandler struct {
	DB *gorm.DB
}

// GetAppProperties handles GET /api/data/app/:document/:collection
// @Summary Get application properties
// @Description Get properties for a specific application document and collection
// @Tags AppData
// @Accept json
// @Produce json
// @Param document path string true "Document ID"
// @Param collection path string true "Collection ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} utils.ErrorResponseStruct
// @Failure 500 {object} utils.ErrorResponseStruct
// @Router /data/app/{document}/{collection} [get]
func (h *AppDataHandler) GetAppProperties(c *fiber.Ctx) error {
	document := c.Params("document")
	collection := c.Params("collection")

	result, err := services.GetApplicationProperties(h.DB, document, collection)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NotFoundResponse(c, fmt.Sprintf("Document '%s' or collection '%s' not found", document, collection))
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "getAppProperties")
	}

	if len(result) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	if !hasContent(result) {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetAppCollectionsAndProperties handles GET /api/data/app/:document?collections=...
// @Summary Get application collections and properties
// @Description Get all collections and properties for a specific application document
// @Tags AppData
// @Accept json
// @Produce json
// @Param document path string true "Document ID"
// @Param collections query string false "Comma-separated list of collections to filter"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} utils.ErrorResponseStruct
// @Failure 500 {object} utils.ErrorResponseStruct
// @Router /data/app/{document} [get]
func (h *AppDataHandler) GetAppCollectionsAndProperties(c *fiber.Ctx) error {
	document := c.Params("document")
	collections := parseCollections(c)

	result, err := services.GetApplicationCollectionsAndProperties(h.DB, document, collections)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NotFoundResponse(c, fmt.Sprintf("Document '%s' not found", document))
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "getAppCollectionsAndProperties")
	}

	if len(result) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	if !hasContent(result) {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetAppDocumentsCollectionsAndProperties handles GET /api/data/app
// @Summary Get all application documents, collections, and properties
// @Description Get all application data
// @Tags AppData
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} utils.ErrorResponseStruct
// @Failure 500 {object} utils.ErrorResponseStruct
// @Router /data/app [get]
func (h *AppDataHandler) GetAppDocumentsCollectionsAndProperties(c *fiber.Ctx) error {
	result, err := services.GetApplicationDocumentsCollectionsAndProperties(h.DB)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NotFoundResponse(c, "No application documents found")
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "getAppDocumentsCollectionsAndProperties")
	}

	if len(result) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	if !hasContent(result) {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// SetAppProperties handles POST /api/data/app/:document
// @Summary Set application properties
// @Description Set properties for a specific application document
// @Tags AppData
// @Accept json
// @Produce json
// @Param document path string true "Document ID"
// @Param body body object true "Properties to set"
// @Success 200 {object} utils.SuccessResponseStruct
// @Failure 400 {object} utils.ErrorResponseStruct
// @Failure 409 {object} utils.ErrorResponseStruct
// @Failure 500 {object} utils.ErrorResponseStruct
// @Router /data/app/{document} [post]
func (h *AppDataHandler) SetAppProperties(c *fiber.Ctx) error {
	document := c.Params("document")

	var body struct {
		Version     types.FlexUint64                         `json:"version"`
		Collections types.FlexList[services.CollectionInput] `json:"collections"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	if document == "" || len(body.Collections) == 0 {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.SetApplicationProperties(h.DB, document, body.Version.Uint64(), body.Collections.Slice())
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "setAppProperties")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}

// DeleteAppCollection handles DELETE /api/data/app/:document/:collection
// @Summary Delete application collection
// @Description Delete a specific collection from an application document
// @Tags AppData
// @Accept json
// @Produce json
// @Param document path string true "Document ID"
// @Param collection path string true "Collection ID"
// @Param body body object true "Version check"
// @Success 200 {object} utils.SuccessResponseStruct
// @Failure 400 {object} utils.ErrorResponseStruct
// @Failure 409 {object} utils.ErrorResponseStruct
// @Failure 500 {object} utils.ErrorResponseStruct
// @Router /data/app/{document}/{collection} [delete]
func (h *AppDataHandler) DeleteAppCollection(c *fiber.Ctx) error {
	document := c.Params("document")
	collection := c.Params("collection")

	var body struct {
		Version types.FlexUint64 `json:"version"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.DeleteApplicationCollection(h.DB, document, body.Version.Uint64(), collection)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "deleteAppCollection")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}

// DeleteAppProperties handles DELETE /api/data/app/:document
// @Summary Delete application properties
// @Description Delete properties from an application document
// @Tags AppData
// @Accept json
// @Produce json
// @Param document path string true "Document ID"
// @Param body body object true "Properties to delete"
// @Success 200 {object} utils.SuccessResponseStruct
// @Failure 400 {object} utils.ErrorResponseStruct
// @Failure 409 {object} utils.ErrorResponseStruct
// @Failure 500 {object} utils.ErrorResponseStruct
// @Router /data/app/{document} [delete]
func (h *AppDataHandler) DeleteAppProperties(c *fiber.Ctx) error {
	document := c.Params("document")

	var body struct {
		Version        types.FlexUint64                               `json:"version"`
		Collections    types.FlexList[services.DeleteCollectionInput] `json:"collections"`
		DeleteDocument bool                                           `json:"deleteDocument"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.DeleteApplicationProperties(h.DB, document, body.Version.Uint64(), body.Collections.Slice(), body.DeleteDocument)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "deleteAppProperties")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}
