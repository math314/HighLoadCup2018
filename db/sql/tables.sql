CREATE TABLE accounts (
  id INT NOT NULL,
  email varchar(100) NOT NULL UNIQUE,
  fname varchar(50),
  sname varchar(50),
  phone varchar(16) UNIQUE, # rare
  sex tinyint(1) NOT NULL,
  birth INT NOT NULL,
  country varchar(50),
  city varchar(50), # rare
  joined INT NOT NULL,
  `status` tinyint(1),
  status_for_recommend tinyint(1),
  premium_start INT NOT NULL,
  premium_end INT NOT NULL,
  premium_now tinyint(1) NOT NULL,

  PRIMARY KEY (`id`)
);

CREATE TABLE interests (
  id INT NOT NULL AUTO_INCREMENT,
  account_id INT NOT NULL,
  interest varchar(100),

  PRIMARY KEY (`id`)
);

CREATE TABLE likes (
  id INT NOT NULL AUTO_INCREMENT,
  account_id_from INT NOT NULL,
  account_id_to INT NOT NULL,
  ts INT NOT NULL,

  PRIMARY KEY (`id`)
);

