package usermanagement

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"siakad/backend/internal/modules/auth"
	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/modules/usermanagement/permissions"
	"siakad/backend/internal/modules/usermanagement/roles"
	"siakad/backend/internal/response"

	"golang.org/x/crypto/bcrypt"
)

type Module struct {
	db                *sql.DB
	auditLog          *auditlogs.Repository
	roleHandler       *roles.Handler
	permissionHandler *permissions.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{db: db}
	if db != nil {
		auditLogRepo := auditlogs.NewRepository(db)
		module.auditLog = auditLogRepo
		module.roleHandler = roles.NewHandler(db, auditLogRepo)
		module.permissionHandler = permissions.NewHandler(db, auditLogRepo)
	}
	return module
}

func (Module) Name() string {
	return "user_management"
}

func (Module) BasePath() string {
	return "/api/v1"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/user-management/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "user management module is ready",
		})
	})

	if m.roleHandler != nil {
		m.roleHandler.RegisterRoutes(mux)
	}
	if m.permissionHandler != nil {
		m.permissionHandler.RegisterRoutes(mux)
	}

	// Assignment endpoints
	if m.db != nil {
		mux.HandleFunc("GET /api/v1/roles/{id}/permissions", m.listRolePermissions)
		mux.HandleFunc("PUT /api/v1/roles/{id}/permissions", m.replaceRolePermissions)
		mux.HandleFunc("GET /api/v1/users/{id}/roles", m.listUserRoles)
		mux.HandleFunc("PUT /api/v1/users/{id}/roles", m.replaceUserRoles)
		mux.HandleFunc("GET /api/v1/users", m.listUsers)
		mux.HandleFunc("POST /api/v1/users", m.createUser)
		mux.HandleFunc("GET /api/v1/users/{id}", m.getUser)
		mux.HandleFunc("PUT /api/v1/users/{id}", m.updateUser)
		mux.HandleFunc("DELETE /api/v1/users/{id}", m.deleteUser)
	}

	if m.roleHandler != nil && m.permissionHandler != nil {
		return
	}

	// Fallback stubs
	mux.HandleFunc("GET /api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/roles/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/roles/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/roles/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/permissions", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/permissions", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/permissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/permissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/permissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}

// --- Role Permissions ---

func (m Module) listRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid role id")
		return
	}

	rows, err := m.db.QueryContext(r.Context(), `
		SELECT p.id, p.name, p.code, COALESCE(p.description, '')
		FROM role_permissions rp
		INNER JOIN permissions p ON p.id = rp.permission_id AND p.deleted_at IS NULL
		WHERE rp.role_id = ?
		ORDER BY p.name ASC
	`, roleID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type permResult struct {
		ID          uint64 `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	items := make([]permResult, 0)
	for rows.Next() {
		var p permResult
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.Description); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": items, "meta": map[string]any{"total": len(items)}})
}

func (m Module) replaceRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req struct {
		PermissionIDs []uint64 `json:"permission_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	tx, err := m.db.BeginTx(r.Context(), nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(r.Context(), "DELETE FROM role_permissions WHERE role_id = ?", roleID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, permID := range req.PermissionIDs {
		if _, err := tx.ExecContext(r.Context(), "INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permID); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var authUserID *uint64
	if user != nil {
		uid := user.UserID
		authUserID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, m.auditLog, "user_management", "replace", "role_permission", roleID, authUserID, req.PermissionIDs)
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "role permissions updated"})
}

// --- User Roles ---

func (m Module) listUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	rows, err := m.db.QueryContext(r.Context(), `
		SELECT r.id, r.name, r.code, COALESCE(r.description, '')
		FROM user_roles ur
		INNER JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
		WHERE ur.user_id = ?
		ORDER BY r.name ASC
	`, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type roleResult struct {
		ID          uint64 `json:"id"`
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	items := make([]roleResult, 0)
	for rows.Next() {
		var rl roleResult
		if err := rows.Scan(&rl.ID, &rl.Name, &rl.Code, &rl.Description); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, rl)
	}
	if err := rows.Err(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": items, "meta": map[string]any{"total": len(items)}})
}

func (m Module) replaceUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		RoleIDs []uint64 `json:"role_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	tx, err := m.db.BeginTx(r.Context(), nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(r.Context(), "DELETE FROM user_roles WHERE user_id = ?", userID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, roleID := range req.RoleIDs {
		if _, err := tx.ExecContext(r.Context(), "INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var authUserID *uint64
	if user != nil {
		uid := user.UserID
		authUserID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, m.auditLog, "user_management", "replace", "user_role", userID, authUserID, req.RoleIDs)
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "user roles updated"})
}

// --- Users CRUD ---

func (m Module) listUsers(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	query := `
		SELECT u.id, u.username, COALESCE(up.full_name, '') AS full_name,
		       COALESCE(u.email, '') AS email, COALESCE(up.phone, '') AS phone,
		       u.is_active, u.created_at
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id AND up.deleted_at IS NULL
		WHERE u.deleted_at IS NULL
	`
	var args []any
	if search != "" {
		query += " AND (u.username LIKE ? OR up.full_name LIKE ? OR u.email LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}
	query += " ORDER BY u.username ASC"

	rows, err := m.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type userResult struct {
		ID        uint64 `json:"id"`
		Username  string `json:"username"`
		FullName  string `json:"full_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		IsActive  bool   `json:"is_active"`
		CreatedAt string `json:"created_at"`
	}
	items := make([]userResult, 0)
	for rows.Next() {
		var u userResult
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.Phone, &u.IsActive, &u.CreatedAt); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, u)
	}
	if err := rows.Err(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": items, "meta": map[string]any{"total": len(items)}})
}

func (m Module) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	fullName := strings.TrimSpace(req.FullName)
	if username == "" {
		response.Error(w, http.StatusBadRequest, "username is required")
		return
	}
	if password == "" {
		response.Error(w, http.StatusBadRequest, "password is required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	tx, err := m.db.BeginTx(r.Context(), nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	var emailPtr, phonePtr any
	if email != "" {
		emailPtr = email
	}
	if phone != "" {
		phonePtr = phone
	}

	result, err := tx.ExecContext(r.Context(), `
		INSERT INTO users (username, password_hash, email, is_active) VALUES (?, ?, ?, ?)
	`, username, string(hash), emailPtr, req.IsActive)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "uk_users_username") {
				response.Error(w, http.StatusConflict, "username already exists")
				return
			}
			response.Error(w, http.StatusConflict, "email already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	userID, _ := result.LastInsertId()

	if fullName != "" {
		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO user_profiles (user_id, full_name, phone) VALUES (?, ?, ?)
		`, userID, fullName, phonePtr); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var authUserID *uint64
	if user != nil {
		uid := user.UserID
		authUserID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, m.auditLog, "user_management", "create", "user", uint64(userID), authUserID, map[string]any{
		"username":  username,
		"full_name": fullName,
		"email":     email,
		"phone":     phone,
		"is_active": req.IsActive,
	})
	m.respondUser(w, r, uint64(userID))
}

func (m Module) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	m.respondUser(w, r, id)
}

func (m Module) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	fullName := strings.TrimSpace(req.FullName)
	if username == "" {
		response.Error(w, http.StatusBadRequest, "username is required")
		return
	}

	tx, err := m.db.BeginTx(r.Context(), nil)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)
	var emailPtr, phonePtr any
	if email != "" {
		emailPtr = email
	}
	if phone != "" {
		phonePtr = phone
	}

	// Update users table
	password := strings.TrimSpace(req.Password)
	if password != "" {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			response.Error(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		_, err = tx.ExecContext(r.Context(), `
			UPDATE users SET username = ?, password_hash = ?, email = ?, is_active = ? WHERE id = ? AND deleted_at IS NULL
		`, username, string(hash), emailPtr, req.IsActive, id)
	} else {
		_, err = tx.ExecContext(r.Context(), `
			UPDATE users SET username = ?, email = ?, is_active = ? WHERE id = ? AND deleted_at IS NULL
		`, username, emailPtr, req.IsActive, id)
	}
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "uk_users_username") {
				response.Error(w, http.StatusConflict, "username already exists")
				return
			}
			response.Error(w, http.StatusConflict, "email already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Upsert user_profiles
	if fullName != "" || phone != "" {
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO user_profiles (user_id, full_name, phone) VALUES (?, ?, ?)
			ON DUPLICATE KEY UPDATE full_name = VALUES(full_name), phone = VALUES(phone)
		`, id, fullName, phonePtr)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var authUserID *uint64
	if user != nil {
		uid := user.UserID
		authUserID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, m.auditLog, "user_management", "update", "user", id, authUserID, map[string]any{
		"username":  username,
		"full_name": fullName,
		"email":     email,
		"phone":     phone,
		"is_active": req.IsActive,
	})
	m.respondUser(w, r, id)
}

func (m Module) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	result, err := m.db.ExecContext(r.Context(), `
		UPDATE users SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}

	// Also soft delete user_profiles
	_, _ = m.db.ExecContext(r.Context(), `
		UPDATE user_profiles SET deleted_at = NOW() WHERE user_id = ?
	`, id)

	user := auth.GetUserFromContext(r.Context())
	var authUserID *uint64
	if user != nil {
		uid := user.UserID
		authUserID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, m.auditLog, "user_management", "delete", "user", id, authUserID, nil)
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "user deleted successfully"})
}

func (m Module) respondUser(w http.ResponseWriter, r *http.Request, id uint64) {
	type userResponse struct {
		ID       uint64 `json:"id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		IsActive bool   `json:"is_active"`
	}

	var u userResponse
	err := m.db.QueryRowContext(r.Context(), `
		SELECT u.id, u.username, COALESCE(up.full_name, '') AS full_name,
		       COALESCE(u.email, '') AS email, COALESCE(up.phone, '') AS phone,
		       u.is_active
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id AND up.deleted_at IS NULL
		WHERE u.id = ? AND u.deleted_at IS NULL
		LIMIT 1
	`, id).Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.Phone, &u.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Load roles
	rows, err := m.db.QueryContext(r.Context(), `
		SELECT r.id, r.name, r.code
		FROM user_roles ur
		INNER JOIN roles r ON r.id = ur.role_id AND r.deleted_at IS NULL
		WHERE ur.user_id = ?
		ORDER BY r.name ASC
	`, id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type roleResult struct {
		ID   uint64 `json:"id"`
		Name string `json:"name"`
		Code string `json:"code"`
	}
	rolesList := make([]roleResult, 0)
	for rows.Next() {
		var rl roleResult
		if err := rows.Scan(&rl.ID, &rl.Name, &rl.Code); err != nil {
			response.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		rolesList = append(rolesList, rl)
	}
	_ = rows.Err()

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"id":        u.ID,
			"username":  u.Username,
			"full_name": u.FullName,
			"email":     u.Email,
			"phone":     u.Phone,
			"is_active": u.IsActive,
			"roles":     rolesList,
		},
	})
}
