-- Seed data for RBAC (Roles, Permissions, Assignments)
-- Run this AFTER schema.sql

-- ============================================================
-- ROLES
-- ============================================================
INSERT INTO roles (name, code, description) VALUES
  ('Administrator', 'admin', 'Full system access with all permissions'),
  ('Academic', 'academic', 'Access to academic module (teachers, subjects, schedules, grades)'),
  ('Student Affairs', 'student_affairs', 'Access to student affairs module (students, enrollments, attendance, discipline)'),
  ('Industry Relations', 'industry_relations', 'Access to industry relations module (companies, internships, alumni)'),
  ('HUBIM', 'hubim', 'Access to HUBIM module (companies, internships, alumni, internship logs)'),
  ('Shared', 'shared', 'Access to shared features (announcements, student search)')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ============================================================
-- PERMISSIONS
-- ============================================================
INSERT INTO permissions (name, code, description) VALUES
  ('Master Read', 'master.read', 'View master data (academic years, semesters, departments, classes, rooms)'),
  ('Master Write', 'master.write', 'Create, update, and delete master data'),
  ('Academic Read', 'academic.read', 'View academic data (teachers, subjects, schedules, assessments, grades)'),
  ('Academic Write', 'academic.write', 'Create, update, and delete academic data'),
  ('Student Affairs Read', 'student_affairs.read', 'View student affairs data (students, enrollments, mutations, graduations, attendance, discipline, extracurricular)'),
  ('Student Affairs Write', 'student_affairs.write', 'Create, update, and delete student affairs data'),
  ('Industry Relations Read', 'industry_relations.read', 'View industry relations data (categories, companies, internships, alumni, internship logs)'),
  ('Industry Relations Write', 'industry_relations.write', 'Create, update, and delete industry relations data'),
  ('Shared Read', 'shared.read', 'View shared features (announcements, student search, audit logs)'),
  ('Shared Write', 'shared.write', 'Create, update, and delete shared features'),
  ('User Management', 'user_management.full', 'Manage roles, permissions, and user-role assignments')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- ============================================================
-- ROLE-PERMISSION ASSIGNMENTS
-- ============================================================

-- Admin gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'admin' AND r.deleted_at IS NULL AND p.deleted_at IS NULL
ON DUPLICATE KEY UPDATE role_id = role_id;

-- Academic gets master + academic read/write
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'academic' AND r.deleted_at IS NULL
  AND p.code IN ('master.read', 'master.write', 'academic.read', 'academic.write')
  AND p.deleted_at IS NULL
ON DUPLICATE KEY UPDATE role_id = role_id;

-- Student Affairs gets student_affairs read/write
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'student_affairs' AND r.deleted_at IS NULL
  AND p.code IN ('student_affairs.read', 'student_affairs.write')
  AND p.deleted_at IS NULL
ON DUPLICATE KEY UPDATE role_id = role_id;

-- Industry Relations gets industry_relations read/write
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'industry_relations' AND r.deleted_at IS NULL
  AND p.code IN ('industry_relations.read', 'industry_relations.write')
  AND p.deleted_at IS NULL
ON DUPLICATE KEY UPDATE role_id = role_id;

-- HUBIM gets industry_relations read/write (same as industry_relations)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'hubim' AND r.deleted_at IS NULL
  AND p.code IN ('industry_relations.read', 'industry_relations.write')
  AND p.deleted_at IS NULL
ON DUPLICATE KEY UPDATE role_id = role_id;

-- Shared gets shared read/write
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'shared' AND r.deleted_at IS NULL
  AND p.code IN ('shared.read', 'shared.write')
  AND p.deleted_at IS NULL
ON DUPLICATE KEY UPDATE role_id = role_id;
