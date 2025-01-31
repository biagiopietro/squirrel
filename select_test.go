package squirrel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelectBuilderToSql(t *testing.T) {
	subQ := Select("aa", "bb").From("dd")
	b := Select("a", "b").
		Prefix("WITH prefix AS ?", 0).
		Distinct().
		Columns("c").
		Column("IF(d IN ("+Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
		Column(Expr("a > ?", 100)).
		Column(Alias(Eq{"b": []int{101, 102, 103}}, "b_alias")).
		Column(Alias(subQ, "subq")).
		From("e").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		Where("f = ?", 4).
		Where(Eq{"g": 5}).
		Where(map[string]interface{}{"h": 6}).
		Where(Eq{"i": []int{7, 8, 9}}).
		Where(Or{Expr("j = ?", 10), And{Eq{"k": 11}, Expr("true")}}).
		GroupBy("l").
		Having("m = n").
		OrderByClause("? DESC", 1).
		OrderBy("o ASC", "p DESC").
		Limit(12).
		Offset(13).
		Suffix("FETCH FIRST ? ROWS ONLY", 14)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"SELECT DISTINCT a, b, c, IF(d IN (?,?,?), 1, 0) as stat_column, a > ?, " +
			"(b IN (?,?,?)) AS b_alias, " +
			"(SELECT aa, bb FROM dd) AS subq " +
			"FROM e " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 " +
			"WHERE f = ? AND g = ? AND h = ? AND i IN (?,?,?) AND (j = ? OR (k = ? AND true)) " +
			"GROUP BY l HAVING m = n ORDER BY ? DESC, o ASC, p DESC LIMIT 12 OFFSET 13 " +
			"FETCH FIRST ? ROWS ONLY"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 100, 101, 102, 103, 4, 5, 6, 7, 8, 9, 10, 11, 1, 14}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderFromSelect(t *testing.T) {
	subQ := Select("c").From("d").Where(Eq{"i": 0})
	b := Select("a", "b").FromSelect(subQ, "subq")
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT a, b FROM (SELECT c FROM d WHERE i = ?) AS subq"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectPageLimitForOracle(t *testing.T) {
	//subQ := Select("c").From("d").Where(Eq{"i": 0})
	subQ := Select("t1.A, t1.B, t2.C, rownum as rnum").
		From("TABLE1 t1").
		Join("TABLE2 t2 ON t1.A = t2.A").
		WhereEscapeEmptyParams(Like{"lower(t2.B)": "f1"}).
		WhereEscapeEmptyParams(Eq{"t2.C": "f2"}).
		WhereEscapeEmptyParams(GtOrEq{"t2.A": "15"}).
		LimitRowNum(2).Page(1)
	sql, args, err := subQ.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT * " +
		"FROM (SELECT t1.A, t1.B, t2.C, rownum as rnum " +
		"FROM TABLE1 t1 JOIN TABLE2 t2 ON t1.A = t2.A " +
		"WHERE lower(t2.B) LIKE ? AND t2.C = ? AND t2.A >= ?) " +
		"WHERE rnum >= 1 AND rnum < 3"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"f1", "f2", "15"}
	assert.Equal(t, expectedArgs, args)
}

func TestWhereEscapeEmptyParams(t *testing.T) {
	//subQ := Select("c").From("d").Where(Eq{"i": 0})
	subQ := Select("t1.A, t1.B, t2.C, rownum as rnum").
		From("TABLE1 t1").
		Join("TABLE2 t2 ON t1.A = t2.A").
		WhereEscapeEmptyParams(Eq{"lower(t2.B)": ""})
	sql, args, err := subQ.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT t1.A, t1.B, t2.C, rownum as rnum " +
		"FROM TABLE1 t1 JOIN TABLE2 t2 ON t1.A = t2.A"
	assert.Equal(t, expectedSql, sql)

	var expectedArgs []interface{}
	assert.Equal(t, expectedArgs, args)
}

func TestWhereEscapeEmptyParams1(t *testing.T) {
	//subQ := Select("c").From("d").Where(Eq{"i": 0})
	subQ := Select("t1.A, t1.B, t2.C, rownum as rnum").
		From("TABLE1 t1").
		Join("TABLE2 t2 ON t1.A = t2.A").
		WhereEscapeEmptyParams(Like{"lower(t2.B)": ""}).
		LimitRowNum(2).Page(1)
	sql, args, err := subQ.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT * " +
		"FROM (SELECT t1.A, t1.B, t2.C, rownum as rnum " +
		"FROM TABLE1 t1 JOIN TABLE2 t2 ON t1.A = t2.A) " +
		"WHERE rnum >= 1 AND rnum < 3"
	assert.Equal(t, expectedSql, sql)

	var expectedArgs []interface{}
	assert.Equal(t, expectedArgs, args)
}

func TestCountAll(t *testing.T) {
	//subQ := Select("c").From("d").Where(Eq{"i": 0})
	subQ := Select("t1.A, t1.B, t2.C, rownum as rnum").
		From("TABLE1 t1").
		Join("TABLE2 t2 ON t1.A = t2.A").
		WhereEscapeEmptyParams(Like{"lower(t2.B)": ""}).CountAll(true)
	sql, args, err := subQ.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT COUNT(*) " +
		"FROM (SELECT t1.A, t1.B, t2.C, rownum as rnum " +
		"FROM TABLE1 t1 JOIN TABLE2 t2 ON t1.A = t2.A)"
	assert.Equal(t, expectedSql, sql)

	var expectedArgs []interface{}
	assert.Equal(t, expectedArgs, args)
}

func TestHasWhereParts(t *testing.T) {
	//subQ := Select("c").From("d").Where(Eq{"i": 0})
	subQ := Select("t1.A, t1.B, t2.C, rownum as rnum").
		From("TABLE1 t1").
		Join("TABLE2 t2 ON t1.A = t2.A").
		WhereEscapeEmptyParams(Like{"lower(t2.B)": ""}).
		Where(Like{"lower(t2.B)": "a"})

	//_,_, err := subQ.ToSql()
	hasWhereParts := subQ.HasWhereParts()
	//assert.NoError(t, err)

	expectedHasWhereParts := true
	assert.Equal(t, expectedHasWhereParts, hasWhereParts)

}

func TestSelectBuilderFromSelectNestedDollarPlaceholders(t *testing.T) {
	subQ := Select("c").
		From("t").
		Where(Gt{"c": 1}).
		PlaceholderFormat(Dollar)
	b := Select("c").
		FromSelect(subQ, "subq").
		Where(Lt{"c": 2}).
		PlaceholderFormat(Dollar)
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT c FROM (SELECT c FROM t WHERE c > $1) AS subq WHERE c < $2"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderToSqlErr(t *testing.T) {
	_, _, err := Select().From("x").ToSql()
	assert.Error(t, err)
}

func TestSelectBuilderPlaceholders(t *testing.T) {
	b := Select("test").Where("x = ? AND y = ?")

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "SELECT test WHERE x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "SELECT test WHERE x = $1 AND y = $2", sql)

	sql, _, _ = b.PlaceholderFormat(Colon).ToSql()
	assert.Equal(t, "SELECT test WHERE x = :1 AND y = :2", sql)
}

func TestSelectBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Select("test").RunWith(db)

	expectedSql := "SELECT test"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.Query()
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRow()
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.Scan()
	assert.NoError(t, err)
}

func TestSelectBuilderNoRunner(t *testing.T) {
	b := Select("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.Query()
	assert.Equal(t, RunnerNotSet, err)

	err = b.Scan()
	assert.Equal(t, RunnerNotSet, err)
}

func TestSelectBuilderSimpleJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN baz ON bar.foo = baz.foo"
	expectedArgs := []interface{}(nil)

	b := Select("*").From("bar").Join("baz ON bar.foo = baz.foo")

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectBuilderParamJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN baz ON bar.foo = baz.foo AND baz.foo = ?"
	expectedArgs := []interface{}{42}

	b := Select("*").From("bar").Join("baz ON bar.foo = baz.foo AND baz.foo = ?", 42)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectBuilderNestedSelectJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN ( SELECT * FROM baz WHERE foo = ? ) r ON bar.foo = r.foo"
	expectedArgs := []interface{}{42}

	nestedSelect := Select("*").From("baz").Where("foo = ?", 42)

	b := Select("*").From("bar").JoinClause(nestedSelect.Prefix("JOIN (").Suffix(") r ON bar.foo = r.foo"))

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectWithOptions(t *testing.T) {
	sql, _, err := Select("*").From("foo").Distinct().Options("SQL_NO_CACHE").ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT SQL_NO_CACHE * FROM foo", sql)
}

func TestSelectWithRemoveLimit(t *testing.T) {
	sql, _, err := Select("*").From("foo").Limit(10).RemoveLimit().ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo", sql)
}

func TestSelectWithRemoveOffset(t *testing.T) {
	sql, _, err := Select("*").From("foo").Offset(10).RemoveOffset().ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo", sql)
}

func TestSelectBuilderNestedSelectDollar(t *testing.T) {
	nestedBuilder := StatementBuilder.PlaceholderFormat(Dollar).Select("*").Prefix("NOT EXISTS (").
		From("bar").Where("y = ?", 42).Suffix(")")
	outerSql, _, err := StatementBuilder.PlaceholderFormat(Dollar).Select("*").
		From("foo").Where("x = ?").Where(nestedBuilder).ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo WHERE x = $1 AND NOT EXISTS ( SELECT * FROM bar WHERE y = $2 )", outerSql)
}

func TestMustSql(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("TestUserFail should have panicked!")
			}
		}()
		// This function should cause a panic
		Select().From("foo").MustSql()
	}()
}

func TestSelectWithoutWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectWithNilWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").Where(nil).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectWithEmptyStringWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").Where("").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}
