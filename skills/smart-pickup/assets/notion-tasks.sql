CREATE TABLE (
  "Task name" TITLE,
  "Status" STATUS('Not started':gray:to_do, 'In progress':blue:in_progress, 'Blocked':red:in_progress, 'Waiting':yellow:in_progress, 'Done':green:complete),
  "Priority" SELECT('High':red, 'Medium':yellow, 'Low':gray),
  "Project" RICH_TEXT,
  "Due date" DATE
)
