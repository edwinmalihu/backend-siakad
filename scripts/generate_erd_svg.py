from __future__ import annotations

from pathlib import Path
from xml.sax.saxutils import escape


OUT_PATH = Path("/Users/edwinmalihu/Documents/internal/SIAKAD/siakad-core-erd.svg")

BOX_W = 300
LINE_H = 24
HEADER_H = 34
BOX_GAP_Y = 28


def box_height(fields: list[str]) -> int:
    return HEADER_H + 16 + len(fields) * LINE_H + 16


entities = {
    "users": {
        "title": "users",
        "group": "Auth",
        "x": 40,
        "y": 80,
        "fields": ["PK id", "username", "password_hash", "email", "is_active"],
    },
    "user_profiles": {
        "title": "user_profiles",
        "group": "Auth",
        "x": 40,
        "y": 320,
        "fields": ["PK id", "FK user_id", "full_name", "phone", "photo_url"],
    },
    "roles": {
        "title": "roles",
        "group": "Auth",
        "x": 40,
        "y": 560,
        "fields": ["PK id", "name", "code"],
    },
    "permissions": {
        "title": "permissions",
        "group": "Auth",
        "x": 40,
        "y": 740,
        "fields": ["PK id", "name", "code"],
    },
    "user_roles": {
        "title": "user_roles",
        "group": "Auth",
        "x": 360,
        "y": 560,
        "fields": ["PK id", "FK user_id", "FK role_id"],
    },
    "role_permissions": {
        "title": "role_permissions",
        "group": "Auth",
        "x": 360,
        "y": 740,
        "fields": ["PK id", "FK role_id", "FK permission_id"],
    },
    "audit_logs": {
        "title": "audit_logs",
        "group": "Shared",
        "x": 40,
        "y": 940,
        "fields": ["PK id", "FK user_id", "module", "action", "entity_type", "entity_id"],
    },
    "announcements": {
        "title": "announcements",
        "group": "Shared",
        "x": 360,
        "y": 980,
        "fields": ["PK id", "FK created_by", "title", "target_scope", "is_published"],
    },
    "academic_years": {
        "title": "academic_years",
        "group": "Master",
        "x": 720,
        "y": 80,
        "fields": ["PK id", "name", "start_date", "end_date", "is_active"],
    },
    "semesters": {
        "title": "semesters",
        "group": "Master",
        "x": 720,
        "y": 320,
        "fields": ["PK id", "FK academic_year_id", "name", "code", "is_active"],
    },
    "departments": {
        "title": "departments",
        "group": "Master",
        "x": 720,
        "y": 560,
        "fields": ["PK id", "code", "name", "program_name"],
    },
    "grade_levels": {
        "title": "grade_levels",
        "group": "Master",
        "x": 720,
        "y": 760,
        "fields": ["PK id", "code", "name", "sort_order"],
    },
    "classes": {
        "title": "classes",
        "group": "Master",
        "x": 1040,
        "y": 560,
        "fields": ["PK id", "FK department_id", "FK grade_level_id", "FK academic_year_id", "name"],
    },
    "rooms": {
        "title": "rooms",
        "group": "Master",
        "x": 1040,
        "y": 820,
        "fields": ["PK id", "code", "name", "type"],
    },
    "students": {
        "title": "students",
        "group": "Student",
        "x": 1400,
        "y": 80,
        "fields": ["PK id", "nis", "nisn", "full_name", "gender", "entry_year", "status"],
    },
    "student_guardians": {
        "title": "student_guardians",
        "group": "Student",
        "x": 1400,
        "y": 360,
        "fields": ["PK id", "FK student_id", "father_name", "mother_name", "guardian_name"],
    },
    "student_enrollments": {
        "title": "student_enrollments",
        "group": "Student",
        "x": 1400,
        "y": 620,
        "fields": ["PK id", "FK student_id", "FK class_id", "FK academic_year_id", "FK semester_id", "status"],
    },
    "student_mutations": {
        "title": "student_mutations",
        "group": "Student",
        "x": 1400,
        "y": 920,
        "fields": ["PK id", "FK student_id", "FK academic_year_id", "FK semester_id", "mutation_type", "status"],
    },
    "student_graduations": {
        "title": "student_graduations",
        "group": "Student",
        "x": 1720,
        "y": 80,
        "fields": ["PK id", "FK student_id", "FK academic_year_id", "graduation_date", "status"],
    },
    "attendances": {
        "title": "attendances",
        "group": "Student",
        "x": 1720,
        "y": 320,
        "fields": ["PK id", "FK student_id", "FK class_id", "attendance_date", "status"],
    },
    "discipline_categories": {
        "title": "discipline_categories",
        "group": "Student",
        "x": 1720,
        "y": 540,
        "fields": ["PK id", "name", "point"],
    },
    "discipline_records": {
        "title": "discipline_records",
        "group": "Student",
        "x": 1720,
        "y": 720,
        "fields": ["PK id", "FK student_id", "FK discipline_category_id", "FK recorded_by", "incident_date"],
    },
    "extracurriculars": {
        "title": "extracurriculars",
        "group": "Student",
        "x": 1720,
        "y": 980,
        "fields": ["PK id", "FK coach_teacher_id", "name", "description"],
    },
    "extracurricular_members": {
        "title": "extracurricular_members",
        "group": "Student",
        "x": 2040,
        "y": 980,
        "fields": ["PK id", "FK extracurricular_id", "FK student_id", "FK academic_year_id", "status"],
    },
    "teachers": {
        "title": "teachers",
        "group": "Academic",
        "x": 2380,
        "y": 80,
        "fields": ["PK id", "nip", "nuptk", "full_name", "employment_status", "position"],
    },
    "homeroom_assignments": {
        "title": "homeroom_assignments",
        "group": "Academic",
        "x": 2380,
        "y": 340,
        "fields": ["PK id", "FK teacher_id", "FK class_id", "FK academic_year_id", "FK semester_id"],
    },
    "subjects": {
        "title": "subjects",
        "group": "Academic",
        "x": 2380,
        "y": 600,
        "fields": ["PK id", "FK department_id", "FK grade_level_id", "code", "name", "kkm"],
    },
    "teaching_devices": {
        "title": "teaching_devices",
        "group": "Academic",
        "x": 2380,
        "y": 880,
        "fields": ["PK id", "FK teacher_id", "FK subject_id", "title", "file_url"],
    },
    "schedules": {
        "title": "schedules",
        "group": "Academic",
        "x": 2700,
        "y": 80,
        "fields": ["PK id", "FK class_id", "FK subject_id", "FK teacher_id", "FK room_id", "FK academic_year_id", "FK semester_id"],
    },
    "assessment_components": {
        "title": "assessment_components",
        "group": "Academic",
        "x": 2700,
        "y": 420,
        "fields": ["PK id", "FK subject_id", "FK academic_year_id", "FK semester_id", "name", "weight"],
    },
    "student_assessments": {
        "title": "student_assessments",
        "group": "Academic",
        "x": 2700,
        "y": 700,
        "fields": ["PK id", "FK student_id", "FK class_id", "FK subject_id", "FK assessment_component_id", "FK teacher_id", "score"],
    },
    "student_grades": {
        "title": "student_grades",
        "group": "Academic",
        "x": 2700,
        "y": 1060,
        "fields": ["PK id", "FK student_id", "FK class_id", "FK subject_id", "FK academic_year_id", "FK semester_id", "final_score"],
    },
    "industry_categories": {
        "title": "industry_categories",
        "group": "Industry",
        "x": 3040,
        "y": 80,
        "fields": ["PK id", "name"],
    },
    "companies": {
        "title": "companies",
        "group": "Industry",
        "x": 3040,
        "y": 260,
        "fields": ["PK id", "FK category_id", "name", "city", "contact_person"],
    },
    "internships": {
        "title": "internships",
        "group": "Industry",
        "x": 3040,
        "y": 500,
        "fields": ["PK id", "FK student_id", "FK company_id", "FK academic_year_id", "start_date", "end_date", "status"],
    },
    "internship_logs": {
        "title": "internship_logs",
        "group": "Industry",
        "x": 3040,
        "y": 840,
        "fields": ["PK id", "FK internship_id", "log_date", "activity"],
    },
    "alumni": {
        "title": "alumni",
        "group": "Industry",
        "x": 3040,
        "y": 1060,
        "fields": ["PK id", "FK student_id", "graduation_year", "current_activity"],
    },
}


relations = [
    ("users", "user_profiles", "1..1"),
    ("users", "user_roles", "1..n"),
    ("roles", "user_roles", "1..n"),
    ("roles", "role_permissions", "1..n"),
    ("permissions", "role_permissions", "1..n"),
    ("users", "audit_logs", "1..n"),
    ("users", "announcements", "1..n"),
    ("academic_years", "semesters", "1..n"),
    ("departments", "classes", "1..n"),
    ("grade_levels", "classes", "1..n"),
    ("academic_years", "classes", "1..n"),
    ("students", "student_guardians", "1..n"),
    ("students", "student_enrollments", "1..n"),
    ("classes", "student_enrollments", "1..n"),
    ("academic_years", "student_enrollments", "1..n"),
    ("semesters", "student_enrollments", "1..n"),
    ("students", "student_mutations", "1..n"),
    ("academic_years", "student_mutations", "1..n"),
    ("semesters", "student_mutations", "1..n"),
    ("students", "student_graduations", "1..n"),
    ("academic_years", "student_graduations", "1..n"),
    ("students", "attendances", "1..n"),
    ("classes", "attendances", "1..n"),
    ("discipline_categories", "discipline_records", "1..n"),
    ("students", "discipline_records", "1..n"),
    ("users", "discipline_records", "1..n"),
    ("teachers", "extracurriculars", "1..n"),
    ("extracurriculars", "extracurricular_members", "1..n"),
    ("students", "extracurricular_members", "1..n"),
    ("academic_years", "extracurricular_members", "1..n"),
    ("teachers", "homeroom_assignments", "1..n"),
    ("classes", "homeroom_assignments", "1..n"),
    ("academic_years", "homeroom_assignments", "1..n"),
    ("semesters", "homeroom_assignments", "1..n"),
    ("departments", "subjects", "1..n"),
    ("grade_levels", "subjects", "1..n"),
    ("teachers", "teaching_devices", "1..n"),
    ("subjects", "teaching_devices", "1..n"),
    ("classes", "schedules", "1..n"),
    ("subjects", "schedules", "1..n"),
    ("teachers", "schedules", "1..n"),
    ("rooms", "schedules", "1..n"),
    ("academic_years", "schedules", "1..n"),
    ("semesters", "schedules", "1..n"),
    ("subjects", "assessment_components", "1..n"),
    ("academic_years", "assessment_components", "1..n"),
    ("semesters", "assessment_components", "1..n"),
    ("students", "student_assessments", "1..n"),
    ("classes", "student_assessments", "1..n"),
    ("subjects", "student_assessments", "1..n"),
    ("assessment_components", "student_assessments", "1..n"),
    ("teachers", "student_assessments", "1..n"),
    ("students", "student_grades", "1..n"),
    ("classes", "student_grades", "1..n"),
    ("subjects", "student_grades", "1..n"),
    ("academic_years", "student_grades", "1..n"),
    ("semesters", "student_grades", "1..n"),
    ("industry_categories", "companies", "1..n"),
    ("students", "internships", "1..n"),
    ("companies", "internships", "1..n"),
    ("academic_years", "internships", "1..n"),
    ("internships", "internship_logs", "1..n"),
    ("students", "alumni", "1..1"),
]


group_colors = {
    "Auth": ("#103b53", "#e9f5fb"),
    "Shared": ("#285943", "#edf9f1"),
    "Master": ("#6a4c0c", "#fff6df"),
    "Student": ("#5a214d", "#fff0fb"),
    "Academic": ("#16486d", "#eef7ff"),
    "Industry": ("#5c3317", "#fff2ea"),
}


def box_geom(key: str) -> tuple[int, int, int, int]:
    entity = entities[key]
    h = box_height(entity["fields"])
    return entity["x"], entity["y"], BOX_W, h


def anchor_right(key: str) -> tuple[int, int]:
    x, y, w, h = box_geom(key)
    return x + w, y + h // 2


def anchor_left(key: str) -> tuple[int, int]:
    x, y, _, h = box_geom(key)
    return x, y + h // 2


def anchor_top(key: str) -> tuple[int, int]:
    x, y, w, _ = box_geom(key)
    return x + w // 2, y


def anchor_bottom(key: str) -> tuple[int, int]:
    x, y, w, h = box_geom(key)
    return x + w // 2, y + h


def route(src: str, dst: str) -> tuple[tuple[int, int], tuple[int, int]]:
    sx, sy, sw, sh = box_geom(src)
    dx, dy, dw, dh = box_geom(dst)
    if sx + sw <= dx:
        return anchor_right(src), anchor_left(dst)
    if dx + dw <= sx:
        return anchor_left(src), anchor_right(dst)
    if sy < dy:
        return anchor_bottom(src), anchor_top(dst)
    return anchor_top(src), anchor_bottom(dst)


svg_parts = []
svg_parts.append(
    '<svg xmlns="http://www.w3.org/2000/svg" width="3380" height="1540" viewBox="0 0 3380 1540">'
)
svg_parts.append(
    "<defs>"
    '<marker id="arrow" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto">'
    '<path d="M0,0 L8,4 L0,8 z" fill="#64748b"/>'
    "</marker>"
    '<style>'
    "text{font-family:Arial,Helvetica,sans-serif;fill:#0f172a}"
    ".title{font-size:28px;font-weight:700}"
    ".subtitle{font-size:13px;fill:#475569}"
    ".header{font-size:17px;font-weight:700;fill:#ffffff}"
    ".field{font-size:13px;fill:#334155}"
    ".label{font-size:11px;fill:#475569}"
    ".legend{font-size:13px;fill:#0f172a}"
    "</style>"
    "</defs>"
)
svg_parts.append('<rect x="0" y="0" width="3380" height="1540" fill="#f8fafc"/>')
svg_parts.append('<text x="40" y="42" class="title">SIAKAD Core ERD</text>')
svg_parts.append(
    '<text x="40" y="66" class="subtitle">Versi inti untuk modular monolith: auth, master, student affairs, academic, industry relations, shared</text>'
)


legend_x = 2040
legend_y = 24
legend_items = ["Auth", "Shared", "Master", "Student", "Academic", "Industry"]
for i, group in enumerate(legend_items):
    dark, light = group_colors[group]
    x = legend_x + i * 210
    svg_parts.append(f'<rect x="{x}" y="{legend_y}" width="18" height="18" rx="4" fill="{dark}"/>')
    svg_parts.append(f'<text x="{x + 28}" y="{legend_y + 14}" class="legend">{group}</text>')


for src, dst, label in relations:
    (x1, y1), (x2, y2) = route(src, dst)
    midx = (x1 + x2) // 2
    svg_parts.append(
        f'<path d="M{x1},{y1} L{midx},{y1} L{midx},{y2} L{x2},{y2}" '
        'fill="none" stroke="#94a3b8" stroke-width="1.6" marker-end="url(#arrow)" opacity="0.9"/>'
    )
    lx = midx + 6
    ly = y1 - 6 if abs(y2 - y1) < 60 else min(y1, y2) + abs(y2 - y1) // 2 - 6
    svg_parts.append(f'<text x="{lx}" y="{ly}" class="label">{escape(label)}</text>')


for key, entity in entities.items():
    x, y, w, h = box_geom(key)
    dark, light = group_colors[entity["group"]]
    svg_parts.append(
        f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="12" fill="{light}" stroke="#cbd5e1" stroke-width="1.5"/>'
    )
    svg_parts.append(
        f'<rect x="{x}" y="{y}" width="{w}" height="{HEADER_H + 8}" rx="12" fill="{dark}"/>'
    )
    svg_parts.append(
        f'<rect x="{x}" y="{y + HEADER_H}" width="{w}" height="8" fill="{dark}"/>'
    )
    svg_parts.append(f'<text x="{x + 16}" y="{y + 24}" class="header">{escape(entity["title"])}</text>')
    fy = y + HEADER_H + 26
    for field in entity["fields"]:
        svg_parts.append(f'<text x="{x + 16}" y="{fy}" class="field">{escape(field)}</text>')
        fy += LINE_H


svg_parts.append("</svg>")

OUT_PATH.write_text("".join(svg_parts), encoding="utf-8")
print(OUT_PATH)
