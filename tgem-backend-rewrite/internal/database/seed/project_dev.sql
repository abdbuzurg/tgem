INSERT INTO projects(name, client, budget, description, signed_date_of_contract, date_start, date_end)
SELECT *
FROM (VALUES
  ('Test Project', 'TGEM', 123456, 'Test Project Description', TO_DATE('2024 Apr 8', 'YYYY Mon DD'), TO_DATE('2024 Apr 8', 'YYYY Mon DD'), TO_DATE('2024 Apr 8', 'YYYY Mon DD')),
  ('Администрирование', 'TGEM', 123456, 'Административная панель', TO_DATE('2024 Apr 8', 'YYYY Mon DD'), TO_DATE('2024 Apr 8', 'YYYY Mon DD'), TO_DATE('2024 Apr 8', 'YYYY Mon DD'))
) AS values_tobe_inserted(name, client, budget, description, signed_date_of_contract, date_start, date_end)
WHERE NOT EXISTS (
  SELECT *
  FROM projects
  WHERE  
    projects.name = values_tobe_inserted.name
    AND projects.client = values_tobe_inserted.client
    AND projects.budget = values_tobe_inserted.budget  
    AND projects.description =  values_tobe_inserted.description 
);

