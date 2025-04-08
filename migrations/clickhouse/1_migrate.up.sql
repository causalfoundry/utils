create table a (
  id String
) 
ENGINE=MergeTree()
ORDER BY id;


create table b (
  id String
) 
ENGINE=MergeTree()
ORDER BY id;
