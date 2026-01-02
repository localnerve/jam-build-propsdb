package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/propsdb/internal/services"
	"github.com/localnerve/propsdb/internal/utils"
	"gorm.io/gorm"
)

// AppDataHandler handles application data routes
type AppDataHandler struct {
	DB *gorm.DB
}

// GetAppProperties handles GET /api/data/app/:document/:collection
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

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetAppCollectionsAndProperties handles GET /api/data/app/:document?collections=...
func (h *AppDataHandler) GetAppCollectionsAndProperties(c *fiber.Ctx) error {
	document := c.Params("document")
	collectionsParam := c.Query("collections", "")

	var collections []string
	if collectionsParam != "" {
		collections = strings.Split(collectionsParam, ",")
	}

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

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetAppDocumentsCollectionsAndProperties handles GET /api/data/app
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

	return c.Status(fiber.StatusOK).JSON(result)
}

// SetAppProperties handles POST /api/data/app/:document
func (h *AppDataHandler) SetAppProperties(c *fiber.Ctx) error {
	document := c.Params("document")

	var body struct {
		Version     uint64                     `json:"version"`
		Collections []services.CollectionInput `json:"collections"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	if document == "" || len(body.Collections) == 0 {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.SetApplicationProperties(h.DB, document, body.Version, body.Collections)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "setAppProperties")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}

// DeleteAppCollection handles DELETE /api/data/app/:document/:collection
func (h *AppDataHandler) DeleteAppCollection(c *fiber.Ctx) error {
	document := c.Params("document")
	collection := c.Params("collection")

	var body struct {
		Version uint64 `json:"version"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.DeleteApplicationCollection(h.DB, document, body.Version, collection)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "deleteAppCollection")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}

// DeleteAppProperties handles DELETE /api/data/app/:document
func (h *AppDataHandler) DeleteAppProperties(c *fiber.Ctx) error {
	document := c.Params("document")

	var body struct {
		Version        uint64                           `json:"version"`
		Collections    []services.DeleteCollectionInput `json:"collections"`
		DeleteDocument bool                             `json:"deleteDocument"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.DeleteApplicationProperties(h.DB, document, body.Version, body.Collections, body.DeleteDocument)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "deleteAppProperties")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}
