-- Story Board: social-style posts with optional media, comments, and message references.

CREATE TABLE IF NOT EXISTS StoryPost (
  ID INT NOT NULL AUTO_INCREMENT,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  author_user_id INT NOT NULL,
  created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (ID),
  KEY ix_story_post_author (author_user_id),
  KEY ix_story_post_created (created_on)
);
