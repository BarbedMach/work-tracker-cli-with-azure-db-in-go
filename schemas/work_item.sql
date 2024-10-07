IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'WorkItem')
BEGIN
    CREATE TABLE WorkItem (
        id INT IDENTITY(1,1) PRIMARY KEY,
        workDate DATE NOT NULL,
        startTime TIME NOT NULL,
        endTime TIME NOT NULL,
        description NVARCHAR(255)
    )
END