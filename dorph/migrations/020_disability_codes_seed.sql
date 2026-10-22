-- Ensure Disability_Codes has rows for student DisabilityCodeId pickers (INSERT IGNORE = safe if already seeded).

INSERT IGNORE INTO Disability_Codes (DBID, DisCodeID, Code, Description) VALUES
  (1, 1, '01', 'Autism spectrum'),
  (1, 2, '02', 'Speech / language'),
  (1, 3, '03', 'OHI'),
  (1, 4, '04', 'Deaf / hard of hearing'),
  (1, 5, '05', 'Visual impairment');
