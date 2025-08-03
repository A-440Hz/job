import React, { useEffect, useState } from 'react';

function formatDate(raw: string) {
  const d = new Date(raw);
  return d.toLocaleDateString();
}

const Item = React.memo(function Item({
  item,
  isEditing,
  isNewItem,
  setIsNewItem,
  editingId,
  setEditingId,
  handleEdit,
  handleDelete,
  handleNew,
}: {
  item: any;
  isEditing: boolean;
  isNewItem: boolean;
  setIsNewItem: (v: boolean) => void;
  editingId: string | null;
  setEditingId: (v: string | null) => void;
  handleEdit: (item: any, title: string, body: string) => void;
  handleDelete: (id: string) => void;
  handleNew: (title: string, body: string) => void;
}) {
  const [title, setTitle] = useState(item.Title);
  const [body, setBody] = useState(item.Body);
  const [saveHighlight, setSaveHighlight] = useState(true);

  useEffect(() => {
    if (item.ID === undefined) {
      setEditingId("new");
    }
  }, [item.ID, setEditingId]);

  const exitEditing = () => {
    setEditingId(null);
    if (item.ID === undefined) {
      setIsNewItem(false);
    } else {
      setTitle(item.Title);
      setBody(item.Body);
    }
  };

  const submitChanges = () => {
    if (item.ID === undefined) {
      if (!title) return; // Prevent blank titles
      handleNew(title, body);
      setIsNewItem(false);
    } else {
      handleEdit(item, title, body);
    }
    setEditingId(null);
  };

  return (
    <>
      {isEditing && <div className="item-modal" onClick={exitEditing} />}
      <li
        className={`py-4 pl-6 pr-6 rounded bg-orange-200 shadow group relative ${
          isEditing ? "item-editing" : ""
        }`}
        onClick={() => (isEditing ? submitChanges() : setEditingId(item.ID || "new"))}
      >
        {isEditing ? (
          <>
          <div onClick={(e) => e.stopPropagation()}>
            <span className="flex items-center justify-between">
              <input
                className="item-title input-box"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="*Company - Position"
                onMouseOver={() => setSaveHighlight(false)}
                onMouseLeave={() => setSaveHighlight(true)}
              />
              {!isNewItem && (
                <button
                  className="rounded-[2vw] text-sm bg-blue-100 border-rose-600 border-2 px-1 text-amber-800 hover:bg-rose-400"
                  onClick={() => handleDelete(item.ID)}
                  onMouseOver={() => setSaveHighlight(false)}
                  onMouseLeave={() => setSaveHighlight(true)}
                >
                  Delete
                </button>
              )}
            </span>
            <textarea
              className="item-body input-box"
              rows={5}
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="notes"
              onMouseOver={() => setSaveHighlight(false)}
              onMouseLeave={() => setSaveHighlight(true)}
            />
          </div>
            <div className="flex items-baseline justify-between">
              {!isNewItem && <p className="text-xs text-gray-900">Created - {formatDate(item.CreatedAt)}</p>}
              <span className="float-right flex border-1 border-amber-500">
                <button
                  className={`mr-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 ${
                    saveHighlight ? "group-hover:bg-emerald-500" : ""
                  }`}
                  onClick={(e) => {
                    e.stopPropagation();
                    submitChanges();
                  }}
                >
                  Save
                </button>
                <button
                  className="ml-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 bg-rose-400 opacity-35 group-hover:bg-orange-200 group-hover:opacity-100 hover:bg-rose-400 hover:opacity-35"
                  onClick={(e) => {
                    e.stopPropagation();
                    exitEditing();
                  }}
                  onMouseOver={() => setSaveHighlight(false)}
                  onMouseLeave={() => setSaveHighlight(true)}
                >
                  Cancel
                </button>
              </span>
            </div>
          </>
        ) : (
          <div className="select-none">
            <span className="flex justify-between items-center">
              <p className="item-title">{item.Title}</p>
              <button
                className="rounded-[2vw] text-sm bg-blue-100 border-blue-300 border-2 px-1 text-amber-800 hover:bg-blue-200"
                onClick={() => setEditingId(item.ID)}
              >
                Edit
              </button>
            </span>
            <p className="item-body">{item.Body}</p>
            <p className="item-timestamp">Created - {formatDate(item.CreatedAt)}</p>
          </div>
        )}
      </li>
    </>
  );
});

export default Item;
