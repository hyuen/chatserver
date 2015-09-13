create table users (
 id serial primary key,
 username varchar(100) UNIQUE not null,
 passwordHash varchar(100) not null,
 passwordSalt varchar(100) not null,
 isDisabled boolean default FALSE
);

create table usersessions (
 SessionKey varchar(100) not null,
 User_id integer not null,
 LoginTime timestamp without time zone default current_timestamp,
 LastSeenTime timestamp without time zone default current_timestamp,
 CONSTRAINT User_child FOREIGN KEY (User_id) REFERENCES Users(id) ON UPDATE CASCADE ON DELETE CASCADE
);
