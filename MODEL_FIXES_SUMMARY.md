# MySQL GORM Model Fixes - Summary

## Problem
GORM AutoMigrate was failing on Railway MySQL with error:
```
Error 1170 (42000): BLOB/TEXT column 'categoryId' used in key specification without a key length
```

MySQL cannot index TEXT/LONGTEXT columns without a length prefix. Any field used in:
- Primary keys
- Unique constraints
- Indexes
- Foreign keys

Must be VARCHAR(191) not TEXT/LONGTEXT.

## Solution
Changed all indexed fields from implicit TEXT to explicit VARCHAR(191).
Removed all `foreignKey` constraint tags (Railway MySQL doesn't support them well).

## Files Modified

### 1. go-backend/models/arena.go
**Changed fields:**
- `QuizCategory.ID` → added `type:varchar(191)`
- `QuizCategory.Slug` → added `type:varchar(191)`
- `Arena.ID` → added `type:varchar(191)`
- `Arena.Slug` → added `type:varchar(191)`
- `Arena.CategoryID` → added `type:varchar(191)`
- `Arena.Category` → removed `foreignKey` tag, changed to `gorm:"-"`
- `Quiz.ID` → added `type:varchar(191)`
- `Quiz.ArenaID` → added `type:varchar(191)`
- `Quiz.CategoryID` → added `type:varchar(191)`
- `Question.ID` → added `type:varchar(191)`
- `Question.QuizID` → added `type:varchar(191)`
- `Question.Options` → removed `foreignKey` tag, changed to `gorm:"-"`
- `Option.ID` → added `type:varchar(191)`
- `Option.QuestionID` → added `type:varchar(191)`

### 2. go-backend/models/user.go
**Changed fields:**
- `User.ID` → added `type:varchar(191)`
- `User.Email` → added `type:varchar(191)`

### 3. go-backend/models/test.go
**Changed fields:**
- `Test.ID` → added `type:varchar(191)`
- `Test.TopicID` → added `type:varchar(191)`
- `Test.CreatedBy` → added `type:varchar(191)`
- `Test.DeletedBy` → added `type:varchar(191)`
- `Test.Creator` → removed `foreignKey` tag, changed to `gorm:"-"`
- `Test.Topic` → removed `foreignKey` tag, changed to `gorm:"-"`
- `Test.Questions` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestQuestion.ID` → added `type:varchar(191)`
- `TestQuestion.TestID` → added `type:varchar(191)`
- `TestQuestion.Test` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestQuestion.MCQOptions` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestQuestion.CodingDetail` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestQuestion.TestCases` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestMCQOption.ID` → added `type:varchar(191)`
- `TestMCQOption.QuestionID` → added `type:varchar(191)`
- `TestCodingDetail.ID` → added `type:varchar(191)`
- `TestCodingDetail.QuestionID` → added `type:varchar(191)`
- `TestCase.ID` → added `type:varchar(191)`
- `TestCase.QuestionID` → added `type:varchar(191)`
- `TestAttempt.ID` → added `type:varchar(191)`
- `TestAttempt.UserID` → added `type:varchar(191)`
- `TestAttempt.TestID` → added `type:varchar(191)`
- `TestAttempt.User` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestAttempt.Test` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestAttempt.Submissions` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestSubmission.ID` → added `type:varchar(191)`
- `TestSubmission.AttemptID` → added `type:varchar(191)`
- `TestSubmission.QuestionID` → added `type:varchar(191)`
- `TestSubmission.SelectedOptionID` → added `type:varchar(191)`
- `TestSubmission.Attempt` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestSubmission.Question` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestResult.ID` → added `type:varchar(191)`
- `TestResult.AttemptID` → added `type:varchar(191)`
- `TestResult.UserID` → added `type:varchar(191)`
- `TestResult.TestID` → added `type:varchar(191)`
- `TestResult.User` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestResult.Test` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestResult.Attempt` → removed `foreignKey` tag, changed to `gorm:"-"`
- `TestViolation.ID` → added `type:varchar(191)`
- `TestViolation.AttemptID` → added `type:varchar(191)`
- `TestViolation.UserID` → added `type:varchar(191)`
- `TestViolation.TestID` → added `type:varchar(191)`

### 4. go-backend/models/attempt.go
**Changed fields:**
- `Attempt.ID` → added `type:varchar(191)`
- `Attempt.UserID` → added `type:varchar(191)`
- `Attempt.QuizID` → added `type:varchar(191)`
- `Attempt.User` → removed `foreignKey` tag, changed to `gorm:"-"`
- `Attempt.Quiz` → removed `foreignKey` tag, changed to `gorm:"-"`
- `AttemptAnswer.ID` → added `type:varchar(191)`
- `AttemptAnswer.AttemptID` → added `type:varchar(191)`
- `AttemptAnswer.QuestionID` → added `type:varchar(191)`
- `AttemptAnswer.SelectedOptionID` → added `type:varchar(191)`

### 5. go-backend/models/topic.go
**Changed fields:**
- `Topic.ID` → added `type:varchar(191)`
- `Topic.Name` → added `type:varchar(191)`
- `Topic.Slug` → added `type:varchar(191)`
- `Topic.CreatedBy` → added `type:varchar(191)`
- `Topic.Creator` → removed `foreignKey` tag, changed to `gorm:"-"`
- `Topic.Tests` → removed `foreignKey` tag, changed to `gorm:"-"`

### 6. go-backend/models/training.go
**Changed fields:**
- `TrainingSession.SessionID` → added `type:varchar(191)`

### 7. go-backend/models/wrong_question.go
**Changed fields:**
- `UserWrongQuestion.ID` → added `type:varchar(191)`
- `UserWrongQuestion.UserID` → added `type:varchar(191)`
- `UserWrongQuestion.AttemptID` → added `type:varchar(191)`
- `UserWrongQuestion.QuestionID` → added `type:varchar(191)`
- `UserWrongQuestion.TestID` → added `type:varchar(191)`
- `UserWrongQuestion.TopicID` → added `type:varchar(191)`
- `UserWrongQuestion.User` → removed `foreignKey` tag, changed to `gorm:"-"`
- `UserWrongQuestion.Question` → removed `foreignKey` tag, changed to `gorm:"-"`
- `UserWrongQuestion.Test` → removed `foreignKey` tag, changed to `gorm:"-"`
- `UserTopicStats.ID` → added `type:varchar(191)`
- `UserTopicStats.UserID` → added `type:varchar(191)`
- `UserTopicStats.TopicID` → added `type:varchar(191)`
- `UserTopicStats.User` → removed `foreignKey` tag, changed to `gorm:"-"`
- `UserTopicStats.Topic` → removed `foreignKey` tag, changed to `gorm:"-"`

## Key Changes

### 1. VARCHAR(191) for All Indexed Fields
Every field that has any of these tags now has `type:varchar(191)`:
- `primaryKey`
- `unique`
- `uniqueIndex`
- `index`

### 2. Removed Foreign Key Constraints
Changed all association fields from:
```go
Category QuizCategory `gorm:"foreignKey:CategoryID" json:"category"`
```

To:
```go
Category QuizCategory `gorm:"-" json:"category"`
```

This removes the foreign key constraint from the database schema while keeping the struct field for JSON serialization. Railway MySQL and PlanetScale don't handle GORM foreign key constraints well.

### 3. Why VARCHAR(191)?
- MySQL with utf8mb4 charset uses 4 bytes per character
- InnoDB has a 767-byte limit for index key prefixes
- 767 / 4 = 191.75, so VARCHAR(191) is the safe maximum
- This is the standard for indexed string fields in MySQL

## Impact

### What Still Works
- All JSON serialization (associations still in structs)
- All queries and joins (manual joins in handlers)
- All existing handler code (no changes needed)
- All frontend code (no changes needed)

### What Changed
- Database schema now uses VARCHAR(191) for IDs and indexed fields
- No foreign key constraints in database (handled in application layer)
- GORM AutoMigrate will now succeed on Railway MySQL

## Testing

After deployment:
1. Check Railway logs for successful migration
2. Verify all tables are created:
   - user
   - quiz_categories
   - arenas
   - quizzes
   - questions
   - options
   - topics
   - tests
   - test_questions
   - test_mcq_options
   - test_coding_details
   - test_cases
   - test_attempts
   - test_submissions
   - test_results
   - test_violations
   - attempts
   - attempt_answers
   - training_questions
   - training_sessions
   - uploads
   - user_wrong_questions
   - user_topic_stats
   - chat_messages

3. Test basic operations:
   - User signup/login
   - Create topic
   - Create test
   - Join arena
   - Submit test
   - View leaderboard
   - Use chat

## Rollback Plan

If issues occur:
1. Revert this commit: `git revert HEAD`
2. Push: `git push origin main`
3. Railway will redeploy with old models

## References

- MySQL Index Length Limits: https://dev.mysql.com/doc/refman/8.0/en/innodb-limits.html
- GORM Data Types: https://gorm.io/docs/models.html#Fields-Tags
- PlanetScale Foreign Keys: https://planetscale.com/docs/learn/operating-without-foreign-key-constraints
