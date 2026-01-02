package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/propsdb/internal/services"
	"github.com/localnerve/propsdb/internal/utils"
	"gorm.io/gorm"
)

// UserDataHandler handles user data routes
type UserDataHandler struct {
	DB *gorm.DB
}

// getUserID extracts user ID from context (set by auth middleware)
func getUserID(c *fiber.Ctx) (string, error) {
	user := c.Locals("user")
	if user == nil {
		return "", fmt.Errorf("user not found in context")
	}

	// The user object from authorizer should have an ID field
	userMap, ok := user.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid user data format")
	}

	userID, ok := userMap["id"].(string)
	if !ok {
		return "", fmt.Errorf("user ID not found")
	}

	return userID, nil
}

// GetUserProperties handles GET /api/data/user/:document/:collection
func (h *UserDataHandler) GetUserProperties(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden, "data.authorization.user")
	}

	document := c.Params("document")
	collection := c.Params("collection")

	result, err := services.GetUserProperties(h.DB, userID, document, collection)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NotFoundResponse(c, fmt.Sprintf("Document '%s' or collection '%s' not found", document, collection))
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "getUserProperties")
	}

	if len(result) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetUserCollectionsAndProperties handles GET /api/data/user/:document?collections=...
func (h *UserDataHandler) GetUserCollectionsAndProperties(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden, "data.authorization.user")
	}

	document := c.Params("document")
	collectionsParam := c.Query("collections", "")

	var collections []string
	if collectionsParam != "" {
		collections = strings.Split(collectionsParam, ",")
	}

	result, err := services.GetUserCollectionsAndProperties(h.DB, userID, document, collections)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NotFoundResponse(c, fmt.Sprintf("Document '%s' not found", document))
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "getUserCollectionsAndProperties")
	}

	if len(result) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// GetUserDocumentsCollectionsAndProperties handles GET /api/data/user
func (h *UserDataHandler) GetUserDocumentsCollectionsAndProperties(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden, "data.authorization.user")
	}

	result, err := services.GetUserDocumentsCollectionsAndProperties(h.DB, userID)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NotFoundResponse(c, "No user documents found")
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "getUserDocumentsCollectionsAndProperties")
	}

	if len(result) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// SetUserProperties handles POST /api/data/user/:document
func (h *UserDataHandler) SetUserProperties(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden, "data.authorization.user")
	}

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

	newVersion, affectedRows, err := services.SetUserProperties(h.DB, userID, document, body.Version, body.Collections)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "setUserProperties")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}

// DeleteUserCollection handles DELETE /api/data/user/:document/:collection
func (h *UserDataHandler) DeleteUserCollection(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden, "data.authorization.user")
	}

	document := c.Params("document")
	collection := c.Params("collection")

	var body struct {
		Version uint64 `json:"version"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.DeleteUserCollection(h.DB, userID, document, body.Version, collection)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "deleteUserCollection")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}

// DeleteUserProperties handles DELETE /api/data/user/:document
func (h *UserDataHandler) DeleteUserProperties(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return utils.ErrorResponse(c, err.Error(), fiber.StatusForbidden, "data.authorization.user")
	}

	document := c.Params("document")

	var body struct {
		Version        uint64                           `json:"version"`
		Collections    []services.DeleteCollectionInput `json:"collections"`
		DeleteDocument bool                             `json:"deleteDocument"`
	}

	if err := c.BodyParser(&body); err != nil {
		return utils.ErrorResponse(c, "Invalid input", fiber.StatusBadRequest, "data.validation.input")
	}

	newVersion, affectedRows, err := services.DeleteUserProperties(h.DB, userID, document, body.Version, body.Collections, body.DeleteDocument)
	if err != nil {
		if strings.Contains(err.Error(), "E_VERSION") {
			return utils.VersionErrorResponse(c)
		}
		return utils.ErrorResponse(c, err.Error(), fiber.StatusInternalServerError, "deleteUserProperties")
	}

	return utils.MutationSuccessResponse(c, newVersion, affectedRows)
}
