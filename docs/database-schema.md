# Database Schema

Eskio uses PostgreSQL 15. The schema is defined in `server/init.sql` and initialized automatically when the Docker container starts.

## Entity Relationship

```
┌──────────┐       ┌──────────────┐       ┌──────────────┐
│  users   │───1:N─│   vouchers   │───1:N─│  line_items  │
└──────────┘       └──────────────┘       └──────┬───────┘
                    │            │                │
                    │ self-ref   │                │
                    │ (corrects/ │          ┌─────┴──────┐
                    │ corrected) │          │  accounts  │
                    └────────────┘          └────────────┘
```

## Tables

### users

| Column        | Type          | Constraints                    |
|---------------|---------------|--------------------------------|
| user_id       | SERIAL        | PRIMARY KEY                    |
| name          | VARCHAR(100)  | NOT NULL                       |
| email         | VARCHAR(255)  | UNIQUE, NOT NULL               |
| password_hash | VARCHAR(255)  | NOT NULL                       |
| role          | VARCHAR(50)   | NOT NULL, DEFAULT 'Bookkeeper' |
| created_at    | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP      |
| updated_at    | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP      |

**Indexes:** `idx_users_email` on email

**Roles:** `Admin`, `Bookkeeper`, `Manager`

---

### accounts

| Column        | Type          | Constraints                               |
|---------------|---------------|-------------------------------------------|
| account_no    | INT           | PRIMARY KEY                               |
| account_name  | VARCHAR(255)  | NOT NULL                                  |
| account_group | INT           | NOT NULL, CHECK (1-8)                     |
| tax_standard  | VARCHAR(50)   | Nullable (e.g. "25%", "12%", "6%", "0%") |
| type          | VARCHAR(10)   | NOT NULL, CHECK ('P&L' or 'BS')          |
| standard_side | VARCHAR(10)   | NOT NULL, CHECK ('Debit' or 'Credit')    |
| created_at    | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP                 |
| updated_at    | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP                 |

**Indexes:** `idx_accounts_group`, `idx_accounts_type`

Uses the Swedish BAS chart of accounts. Accounts are pre-seeded via `init.sql` and `bas_accounts_*.sql`.

---

### vouchers

| Column                  | Type          | Constraints                         |
|-------------------------|---------------|-------------------------------------|
| voucher_id              | SERIAL        | PRIMARY KEY                         |
| voucher_number          | INTEGER       | UNIQUE, auto-increment via sequence |
| date                    | DATE          | NOT NULL                            |
| description             | TEXT          | NOT NULL                            |
| reference               | VARCHAR(255)  | Nullable                            |
| total_amount            | DECIMAL(15,2) | NOT NULL                            |
| period                  | VARCHAR(7)    | NOT NULL (format: YYYY-MM)          |
| created_by              | INT           | FK -> users, ON DELETE RESTRICT     |
| corrects_voucher_id     | INT           | FK -> vouchers, ON DELETE SET NULL  |
| corrected_by_voucher_id | INT           | FK -> vouchers, ON DELETE SET NULL  |
| created_at              | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP           |
| updated_at              | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP           |

**Indexes:** `idx_vouchers_period`, `idx_vouchers_created_by`, `idx_vouchers_date`, `idx_vouchers_number`

**Correction model:** When voucher A is corrected by voucher B:
- A.corrected_by_voucher_id = B.voucher_id
- B.corrects_voucher_id = A.voucher_id

---

### line_items

| Column        | Type          | Constraints                                        |
|---------------|---------------|----------------------------------------------------|
| line_id       | SERIAL        | PRIMARY KEY                                        |
| voucher_id    | INT           | NOT NULL, FK -> vouchers, ON DELETE CASCADE        |
| account_no    | INT           | NOT NULL, FK -> accounts, ON DELETE RESTRICT       |
| debit_amount  | DECIMAL(15,2) | NOT NULL, DEFAULT 0                                |
| credit_amount | DECIMAL(15,2) | NOT NULL, DEFAULT 0                                |
| tax_code      | INT           | Nullable (0, 6, 12, or 25)                        |
| project_id    | INT           | Nullable                                           |
| cost_center_id| INT           | Nullable                                           |
| created_at    | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP                          |
| updated_at    | TIMESTAMP     | DEFAULT CURRENT_TIMESTAMP                          |

**Indexes:** `idx_line_items_voucher`, `idx_line_items_account`

**Check constraint:** Exactly one of `debit_amount` or `credit_amount` must be > 0.

```sql
CHECK (
    (debit_amount > 0 AND credit_amount = 0) OR
    (credit_amount > 0 AND debit_amount = 0)
)
```

### scheduled_tasks

| Column              | Type      | Constraints                                    |
|---------------------|-----------|------------------------------------------------|
| task_id             | SERIAL    | PRIMARY KEY                                    |
| user_id             | INT       | NOT NULL, FK -> users, ON DELETE CASCADE       |
| description         | TEXT      | NOT NULL                                       |
| prompt              | TEXT      | NOT NULL                                       |
| template_voucher_id | INT       | FK -> vouchers, ON DELETE SET NULL             |
| day_of_month        | INT       | NOT NULL, CHECK (1-31)                         |
| active              | BOOLEAN   | NOT NULL, DEFAULT true                         |
| last_run_at         | TIMESTAMP | Nullable                                       |
| next_run_at         | TIMESTAMP | Nullable                                       |
| created_at          | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP                      |
| updated_at          | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP                      |

**Indexes:** `idx_scheduled_tasks_user`, `idx_scheduled_tasks_next_run`, `idx_scheduled_tasks_active`

Used by the AI agent's scheduler to execute recurring voucher creation.

---

### agent_messages

| Column          | Type        | Constraints                              |
|-----------------|-------------|------------------------------------------|
| message_id      | SERIAL      | PRIMARY KEY                              |
| user_id         | INT         | NOT NULL, FK -> users, ON DELETE CASCADE |
| conversation_id | VARCHAR(36) | NOT NULL                                 |
| role            | VARCHAR(20) | NOT NULL, CHECK ('user' or 'assistant')  |
| content         | TEXT        | NOT NULL                                 |
| created_at      | TIMESTAMP   | DEFAULT CURRENT_TIMESTAMP                |

**Indexes:** `idx_agent_messages_user`, `idx_agent_messages_conversation`

Stores chat history between users and the AI assistant.

---

## Key Relationships

- A **user** can create many **vouchers** (`created_by` FK)
- A **voucher** has many **line items** (cascade delete)
- Each **line item** references one **account** (restrict delete)
- A **voucher** can correct another voucher (self-referencing FK pair)
- A **user** can have many **scheduled tasks** (cascade delete)
- A **scheduled task** can reference a template **voucher** (set null on delete)
- A **user** can have many **agent messages** (cascade delete)
- Deleting a user is restricted if they have vouchers
- Deleting an account is restricted if it has line items
