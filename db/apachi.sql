
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  UserID int(11) NOT NULL AUTO_INCREMENT,
  Username varchar(51) NOT NULL,
  Hash varchar(60) NOT NULL,
  UserType varchar(45) NOT NULL,
  Secret varchar(45) DEFAULT NULL,
  PRIMARY KEY (`UserID`)
) 

CREATE TABLE log (
  `Id` INT NOT NULL AUTO_INCREMENT,
  `logMsg` VARCHAR(45) NOT NULL,
  `FromUserId` VARCHAR(45) NOT NULL,
  `Value` GLOB NULL,
  PRIMARY KEY (`Id`));


DROP ROLE IF EXISTS my_user;
CREATE ROLE my_user LOGIN PASSWORD 'my_password';
GRANT INSERT, SELECT TO my_user;

