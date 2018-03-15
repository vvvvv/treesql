package treesql

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/vilterp/treesql/pkg/lang"
	"github.com/vilterp/treesql/pkg/util"
)

func TestLangExec(t *testing.T) {
	tsr := runSimpleTestScript(t, []simpleTestStmt{
		// TODO: maybe dedup this with SelectTest?
		{
			stmt: `
				CREATETABLE blog_posts (
					id string PRIMARYKEY,
					title string
				)
			`,
			ack: "CREATE TABLE",
		},
		{
			stmt: `
				CREATETABLE comments (
					id string PRIMARYKEY,
					blog_post_id string REFERENCESTABLE blog_posts,
					body string
				)
			`,
			ack: "CREATE TABLE",
		},
		// Insert data.
		{
			stmt: `INSERT INTO blog_posts VALUES ("0", "hello world")`,
			ack:  "INSERT 1",
		},
		{
			stmt: `INSERT INTO blog_posts VALUES ("1", "hello again world")`,
			ack:  "INSERT 1",
		},
		{
			stmt: `INSERT INTO comments VALUES ("0", "0", "hello yourself!")`,
			ack:  "INSERT 1",
		},
		{
			stmt: `INSERT INTO comments VALUES ("1", "1", "sup")`,
			ack:  "INSERT 1",
		},
		{
			stmt: `INSERT INTO comments VALUES ("2", "1", "so creative")`,
			ack:  "INSERT 1",
		},
	})
	defer tsr.Close()

	db := tsr.server.db

	// Common stuff
	scanPostsByID := lang.NewMemberAccess(
		lang.NewMemberAccess(lang.NewVar("blog_posts"), "id"),
		"scan",
	)
	blogPostType := db.schema.tables["blog_posts"].getType()

	// Cases
	testCases := []struct {
		in         lang.Expr
		prettyExpr string
		typ        string
		outJSON    string
	}{
		{
			scanPostsByID,
			`blog_posts.id.scan`,
			`Iterator<{
  id: string,
  title: string,
}>`,
			`[
					{"id": "0", "title": "hello world"},
					{"id": "1", "title": "hello again world"}
			]`,
		},
		{
			lang.NewFuncCall("map", []lang.Expr{
				scanPostsByID,
				lang.NewELambda(
					[]lang.Param{{"post", blogPostType}},
					lang.NewMemberAccess(lang.NewVar("post"), "title"),
					lang.TString,
				),
			}),
			`map(blog_posts.id.scan, (post: {
  id: string,
  title: string,
}): string => post.title)`,
			`Iterator<string>`,
			`["hello world", "hello again world"]`,
		},
	}

	for idx, testCase := range testCases {
		// Construct transaction.
		boltTxn, err := db.boltDB.Begin(false)
		if err != nil {
			t.Fatal(err)
		}
		txn := &txn{
			boltTxn: boltTxn,
			db:      db,
		}

		// Check pretty printed form.
		pretty := testCase.in.Format().String()
		if pretty != testCase.prettyExpr {
			t.Errorf("case %d: expected pretty form\n\n%s\n\ngot\n\n%s", idx, testCase.prettyExpr, pretty)
			continue
		}

		// Construct scope.
		userRootScope, _ := db.schema.toScope(txn)

		// Get type; compare.
		// TODO: unfortuantely this is pretty messed up.
		// probably need to rethink type scopes.
		//typ, err := testCase.in.GetType(typeScope)
		//if err != nil {
		//	t.Errorf("case %d: %v", idx, err)
		//	continue
		//}
		//if typ.Format().String() != testCase.typ {
		//	t.Errorf("case %d: expected %s; got %s", idx, testCase.typ, typ.Format())
		//	continue
		//}

		// Interpret the test expression.
		interp := lang.NewInterpreter(userRootScope, testCase.in)
		val, err := interp.Interpret()
		if err != nil {
			// TODO: test for error
			t.Errorf("case %d: %v", idx, err)
			continue
		}

		// Get the output as a string of JSON.
		buf := bytes.NewBufferString("")
		bufWriter := bufio.NewWriter(buf)
		if err := val.WriteAsJSON(bufWriter, interp); err != nil {
			t.Errorf("case %d: %v", idx, err)
			continue
		}
		bufWriter.Flush()
		json := buf.String()

		// Compare expected and actual JSON.
		eq, err := util.AreEqualJSON(json, testCase.outJSON)
		if err != nil {
			t.Errorf(`case %d: %v`, idx, err)
			continue
		}
		if !eq {
			t.Errorf("case %d: EXPECTED\n\n%s\n\nGOT:\n\n%s\n", idx, testCase.outJSON, json)
		}
	}
}