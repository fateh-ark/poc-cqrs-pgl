-- CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'replicator_password';
-- SELECT pg_create_physical_replication_slot('replication_slot');

create schema test_schema;

create table test_schema.books
(
    id     serial
        constraint books_pk
            primary key,
    title  varchar(128),
    author varchar(128)
);

insert into test_schema.books (title, author) values ('Parisienne, La (Une parisienne)', 'Billy Toppes');
insert into test_schema.books (title, author) values ('Ginger Snaps', 'Kalie Nashe');
insert into test_schema.books (title, author) values ('Bloodsuckers', 'Ruth Browell');
insert into test_schema.books (title, author) values ('Berlin Express', 'April Dolman');
insert into test_schema.books (title, author) values ('Battle of Shaker Heights, The', 'Rayna Marrable');
insert into test_schema.books (title, author) values ('Timecop', 'Ally Groucock');
insert into test_schema.books (title, author) values ('Into the Storm', 'Avram Kerswill');
insert into test_schema.books (title, author) values ('Nadine', 'Jackquelin Stowe');
insert into test_schema.books (title, author) values ('Temptation (Tentação)', 'Herold Tyers');
insert into test_schema.books (title, author) values ('Block Party (a.k.a. Dave Chappelle''s Block Party)', 'Xymenes Feare');
insert into test_schema.books (title, author) values ('King Lear (Korol Lir)', 'Gard Le feuvre');
insert into test_schema.books (title, author) values ('Shock', 'Arnold Van der Hoven');
insert into test_schema.books (title, author) values ('The Horseplayer', 'Caddric Innocenti');
insert into test_schema.books (title, author) values ('Female', 'Jule Ellin');
insert into test_schema.books (title, author) values ('Rio 2', 'Doris Barkus');
