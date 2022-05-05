
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  UserID int(11) NOT NULL AUTO_INCREMENT,
  Username varchar(51) NOT NULL,
  Hash varchar(60) NOT NULL,
  UserType varchar(45) NOT NULL,
  Secret varchar(45) DEFAULT NULL,
  PRIMARY KEY (`UserID`)
) 
