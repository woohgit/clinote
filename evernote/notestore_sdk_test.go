package evernote

import (
	"errors"
	"testing"

	"github.com/TcM1911/evernote-sdk-golang/notestore"
	"github.com/TcM1911/evernote-sdk-golang/types"
	"github.com/stretchr/testify/assert"
)

var errExpected = errors.New("expected")

func TestGetAllNotebooks(t *testing.T) {
	assert := assert.New(t)
	token := "token"
	c := &mockClient{apiToken: token}
	t.Run("Return error from api", func(t *testing.T) {
		api := &mockAPI{listNotebooks: func(key string) ([]*types.Notebook, error) { return nil, errExpected }}
		ns := &Notestore{client: c, evernoteNS: api}
		books, err := ns.GetAllNotebooks()
		assert.Nil(books, "No notebooks should be returned")
		assert.Equal(errExpected, err, "Wrong error returned")
	})
	t.Run("Return all books", func(t *testing.T) {
		title := "Name"
		expectedBooks := []*Notebook{&Notebook{Name: title}}
		books := []*types.Notebook{&types.Notebook{Name: &title}}
		api := &mockAPI{listNotebooks: func(key string) ([]*types.Notebook, error) { return books, nil }}
		ns := &Notestore{client: c, evernoteNS: api}
		bs, err := ns.GetAllNotebooks()
		assert.Equal(expectedBooks, bs, "Notebooks should be returned")
		assert.NoError(err, "No error returned")
	})
}

func TestUpdateNotebookSDK(t *testing.T) {
	assert := assert.New(t)
	token := "token"
	c := &mockClient{apiToken: token}
	guid := "guid"
	t.Run("Return ErrNoNotebookCached", func(t *testing.T) {
		ns := &Notestore{client: c, evernoteNS: nil}
		err := ns.UpdateNotebook(&Notebook{})
		assert.Equal(ErrNoNotebookCached, err, "No cached notebooks")
	})
	t.Run("Return ErrNoNotebookFound", func(t *testing.T) {
		cachedGUID := types.GUID("another guid")
		cachedNB := &types.Notebook{GUID: &cachedGUID}
		notCached := "not cached"
		cacheNotebook(cachedNB)
		ns := &Notestore{client: c, evernoteNS: nil}
		err := ns.UpdateNotebook(&Notebook{GUID: notCached})
		assert.Equal(ErrNoNotebookFound, err, "Notebook not cached")
	})
	t.Run("Return error from api", func(t *testing.T) {
		oldTitle := "Old title"
		newTitle := "New title"
		savedGUID := types.GUID(guid)
		cachedNotebook := &types.Notebook{GUID: &savedGUID, Name: &oldTitle}
		cacheNotebook(cachedNotebook)
		api := &mockAPI{updateNotebook: func(k string, nb *types.Notebook) (int32, error) { return int32(0), errExpected }}
		ns := &Notestore{client: c, evernoteNS: api}
		book := &Notebook{GUID: guid, Name: newTitle}
		err := ns.UpdateNotebook(book)
		assert.Error(err, "Should return error from api call")
	})
	t.Run("Update notebook", func(t *testing.T) {
		oldTitle := "Old title"
		newTitle := "New title"
		savedGUID := types.GUID(guid)
		cachedNotebook := &types.Notebook{GUID: &savedGUID, Name: &oldTitle}
		cacheNotebook(cachedNotebook)
		var saved *types.Notebook
		api := &mockAPI{updateNotebook: func(k string, nb *types.Notebook) (int32, error) { saved = nb; return int32(0), nil }}
		ns := &Notestore{client: c, evernoteNS: api}
		book := &Notebook{GUID: guid, Name: newTitle}
		err := ns.UpdateNotebook(book)
		assert.NoError(err, "Should update without error")
		assert.Equal(newTitle, *saved.Name, "Should update notebook name")
	})
}

func TestCreateNotebookSDK(t *testing.T) {
	assert := assert.New(t)
	token := "token"
	c := &mockClient{apiToken: token}
	var saved *types.Notebook
	name := "Notebook name"
	stack := "Stack name"
	nb := &Notebook{Name: name, Stack: stack}
	api := &mockAPI{createNotebook: func(k string, nb *types.Notebook) (*types.Notebook, error) { saved = nb; return nil, errExpected }}
	ns := &Notestore{client: c, evernoteNS: api}
	err := ns.CreateNotebook(nb, false)
	assert.Equal(errExpected, err, "Wrong error returned")
	assert.Equal(name, *saved.Name, "Wrong notebook name")
	assert.Equal(stack, *saved.Stack, "Wrong stack")
}

func TestCreateNoteSDK(t *testing.T) {
	assert := assert.New(t)
	token := "token"
	c := &mockClient{apiToken: token}
	var saved *types.Note
	notebookGUID := "Some GUID"
	note := &Note{
		Notebook: &Notebook{GUID: notebookGUID, Name: "Name"},
		Title:    "Note title",
		Body:     "Note body",
	}
	ns := &Notestore{
		client:     c,
		evernoteNS: &mockAPI{createNote: func(k string, n *types.Note) (*types.Note, error) { saved = n; return nil, errExpected }},
	}
	err := ns.CreateNote(note)
	assert.Equal(errExpected, err, "Wrong error")
	assert.Equal(&note.Body, saved.Content, "Body not saved")
	assert.Equal(&note.Title, saved.Title, "Title not saved")
	assert.Equal(notebookGUID, *saved.NotebookGuid, "Notebook GUID doesn't match")
}

func TestDeleteNoteSDK(t *testing.T) {
	assert := assert.New(t)
	token := "token"
	c := &mockClient{apiToken: token}
	notebookGUID := "Some GUID"
	ns := &Notestore{
		client:     c,
		evernoteNS: &mockAPI{deleteNote: func(a string, g types.GUID) (int32, error) { return int32(0), nil }},
	}

	err := ns.DeleteNote(notebookGUID)
	assert.NoError(err, "Should not return an error.")
}

func TestUpdateNoteSDK(t *testing.T) {
	assert := assert.New(t)
	token := "token"
	c := &mockClient{apiToken: token}
	ns := &Notestore{
		client: c,
	}

	t.Run("error when no GUID", func(t *testing.T) {
		err := ns.UpdateNote(&Note{})
		assert.Equal(ErrNoGUIDSet, err, "Wrong error returned")
	})

	t.Run("error when no title", func(t *testing.T) {
		err := ns.UpdateNote(&Note{GUID: "some guid"})
		assert.Equal(ErrNoTitleSet, err, "Wrong error returned")
	})

	t.Run("Skip body if empty", func(t *testing.T) {
		var expectedNote *types.Note
		expectedGUID := "Expected GUID"
		expectedTitle := "Expected Title"
		ns.evernoteNS = &mockAPI{updateNote: func(api string, n *types.Note) (*types.Note, error) { expectedNote = n; return nil, nil }}
		err := ns.UpdateNote(&Note{
			Title: expectedTitle,
			GUID:  expectedGUID,
		})
		assert.NoError(err, "No error should be returned")
		assert.Equal(expectedGUID, string(expectedNote.GetGUID()), "Wrong GUID")
		assert.Equal(expectedTitle, expectedNote.GetTitle(), "Wrong Title")
		assert.Equal("", expectedNote.GetContent(), "Content should be empty")
	})

	t.Run("Include body if set", func(t *testing.T) {
		var expectedNote *types.Note
		expectedGUID := "Expected GUID"
		expectedTitle := "Expected Title"
		expectedContent := "This is note content"
		ns.evernoteNS = &mockAPI{updateNote: func(api string, n *types.Note) (*types.Note, error) { expectedNote = n; return nil, nil }}
		err := ns.UpdateNote(&Note{
			Title: expectedTitle,
			GUID:  expectedGUID,
			Body:  expectedContent,
		})
		assert.NoError(err, "No error should be returned")
		assert.Equal(expectedGUID, string(expectedNote.GetGUID()), "Wrong GUID")
		assert.Equal(expectedTitle, expectedNote.GetTitle(), "Wrong Title")
		assert.Equal(expectedContent, expectedNote.GetContent(), "Content should be empty")
	})
}

func TestFindNotes(t *testing.T) {
	assert := assert.New(t)
	expectedNote := types.NewNote()
	GUID := types.GUID("Note GUID")
	title := "Note title"
	expectedNote.GUID = &GUID
	expectedNote.Title = &title
	nl := &notestore.NoteList{Notes: []*types.Note{expectedNote}}
	token := "token"
	c := &mockClient{apiToken: token}
	ns := &Notestore{
		client:     c,
		evernoteNS: &mockAPI{findNote: func(string, *notestore.NoteFilter, int32, int32) (*notestore.NoteList, error) { return nl, nil }},
	}
	t.Run("all notebooks", func(t *testing.T) {
		filter := &NoteFilter{Words: "search term"}
		notes, err := ns.FindNotes(filter, 0, 20)
		assert.NoError(err, "Should not return an error")
		assert.Len(notes, 1, "Wrong number of notes returned.")
		assert.Equal(title, notes[0].Title, "Wrong title")
		assert.Equal(string(GUID), notes[0].GUID, "Wrong GUID")
	})

	t.Run("one notebook", func(t *testing.T) {
		filter := &NoteFilter{NotebookGUID: "Book GUID"}
		notes, err := ns.FindNotes(filter, 0, 20)
		assert.NoError(err, "Should not return an error")
		assert.Len(notes, 1, "Wrong number of notes returned.")
		assert.Equal(title, notes[0].Title, "Wrong title")
		assert.Equal(string(GUID), notes[0].GUID, "Wrong GUID")
	})

	t.Run("return error", func(t *testing.T) {
		filter := &NoteFilter{NotebookGUID: "Book GUID"}
		expectedErr := errors.New("expected")
		ns.evernoteNS = &mockAPI{findNote: func(string, *notestore.NoteFilter, int32, int32) (*notestore.NoteList, error) {
			return nil, expectedErr
		}}
		notes, err := ns.FindNotes(filter, 0, 20)
		assert.Error(err, "Should return an error")
		assert.Nil(notes, "Notes should be nil")
		assert.Equal(expectedErr, err, "Wrong error")
	})
}

func TestGetNoteContentSDK(t *testing.T) {
	assert := assert.New(t)
	expectedContent := "Note content"
	c := &mockClient{apiToken: "token"}
	ns := &Notestore{
		client:     c,
		evernoteNS: &mockAPI{getNoteContent: func(string, types.GUID) (string, error) { return expectedContent, nil }},
	}
	content, err := ns.GetNoteContent("GUID")
	assert.NoError(err, "No error should be returned")
	assert.Equal(expectedContent, content, "Wrong content")
}

type mockAPI struct {
	listNotebooks  func(string) ([]*types.Notebook, error)
	updateNotebook func(string, *types.Notebook) (int32, error)
	createNotebook func(string, *types.Notebook) (*types.Notebook, error)
	createNote     func(string, *types.Note) (*types.Note, error)
	deleteNote     func(string, types.GUID) (int32, error)
	updateNote     func(string, *types.Note) (*types.Note, error)
	findNote       func(string, *notestore.NoteFilter, int32, int32) (*notestore.NoteList, error)
	getNoteContent func(string, types.GUID) (string, error)
}

func (a *mockAPI) ListNotebooks(apiKey string) (r []*types.Notebook, err error) {
	return a.listNotebooks(apiKey)
}

func (a *mockAPI) CreateNotebook(apiKey string, notebook *types.Notebook) (r *types.Notebook, err error) {
	return a.createNotebook(apiKey, notebook)
}

func (a *mockAPI) UpdateNotebook(apiKey string, notebook *types.Notebook) (r int32, err error) {
	return a.updateNotebook(apiKey, notebook)
}

func (a *mockAPI) CreateNote(apiKey string, note *types.Note) (r *types.Note, err error) {
	return a.createNote(apiKey, note)
}

func (a *mockAPI) FindNotes(apiKey string, filter *notestore.NoteFilter, offset int32, maxNumNotes int32) (r *notestore.NoteList, err error) {
	return a.findNote(apiKey, filter, offset, maxNumNotes)
}

func (a *mockAPI) DeleteNote(apiKey string, guid types.GUID) (int32, error) {
	return a.deleteNote(apiKey, guid)
}

func (a *mockAPI) UpdateNote(authenticationToken string, note *types.Note) (r *types.Note, err error) {
	return a.updateNote(authenticationToken, note)
}

func (a *mockAPI) GetNoteContent(authenticationToken string, guid types.GUID) (r string, err error) {
	return a.getNoteContent(authenticationToken, guid)
}

func (a *mockAPI) GetNotebook(authenticationToken string, guid types.GUID) (r *types.Notebook, err error) {
	panic("not implemented")
}
