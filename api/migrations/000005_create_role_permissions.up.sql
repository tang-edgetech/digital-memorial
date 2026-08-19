CREATE TABLE role_permissions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  role ENUM('admin','agent') NOT NULL,
  module VARCHAR(50) NOT NULL,
  action VARCHAR(50) NOT NULL,
  allowed BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  updated_by BIGINT UNSIGNED NULL,
  UNIQUE KEY uq_role_module_action (role, module, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO role_permissions (role, module, action, allowed) VALUES
  ('admin','users','view',true),
  ('admin','users','create',true),
  ('admin','users','edit',true),
  ('admin','users','delete',true),
  ('admin','users','enable_disable',true),
  ('admin','settings','edit',true),
  ('agent','users','view',true),
  ('agent','users','create',false),
  ('agent','users','edit',false),
  ('agent','users','delete',false),
  ('agent','users','enable_disable',false),
  ('agent','settings','edit',false);
