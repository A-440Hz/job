import React, { useEffect, useState } from 'react';
import { formatDate } from './api/datetime';

const Item = React.memo(function Item({
  item,
  isEditing,
  isNewItem,
  setIsNewItem,
  setEditingId,
  handleEdit,
  handleDelete,
  handleNew,
}: {
  item: any;
  isEditing: boolean;
  isNewItem: boolean;
  setIsNewItem: (v: boolean) => void;
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

  useEffect(() => {
  if (!isEditing) {
    setTitle(item.Title);
    setBody(item.Body);
  }
}, [item.Title, item.Body, isEditing]);

  const exitEditing = () => {
    setEditingId(null);
    if (item.ID === undefined) {
      setIsNewItem(false);
    } else {
      // setTitle(item.Title);
      // setBody(item.Body);
    }
  };

  const submitChanges = () => {
    if (item.ID === undefined) {
      if (!title) return; // Prevent blank titles. TODO:a flashing animation for the title input border 
      handleNew(title, body);
    } else {
      handleEdit(item, title, body);
    }
    exitEditing();
  };

  return (
    <>
      {isEditing && <div className="item-modal" onClick={exitEditing} />}
      <li
        className={`py-4 px-6 rounded bg-amber-100 shadow group relative ${
          isEditing ? "item-editing" : ""
        } ${
        saveHighlight? "hover:border-emerald-500" : "hover:border-blue-200"
        } border-rose-400`}
        onClick={() => (isEditing ? submitChanges() : setEditingId(item.ID || "new"))}
      >
        {isEditing ? (
          <>
          <div onClick={(e) => e.stopPropagation()}>
            <span className="flex items-start justify-between">
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
                  className="delete rounded-[2vw] text-sm bg-orange-50 border-rose-400 border-2 px-1 text-red-700 hover:bg-rose-400 hover:scale-96 transition-all"
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
              {!isNewItem && <p className="text-xs text-gray-900">Created - {formatDate(item.CreatedAt).toLocaleDateString()}</p>}
              <span className="float-right flex">
                <button
                  className={`mr-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 border-emerald-600 text-emerald-600 bg-orange-50 hover:scale-96 transition-all ${
                    saveHighlight ? "group-hover:bg-emerald-500 group-hover:opacity-75 group-hover:scale-96" : ""
                  }`}
                  onClick={(e) => {
                    e.stopPropagation();
                    submitChanges();
                  }}
                >
                  Save
                </button>
                <button
                  className="cancel ml-2 rounded-[2vw] text-sm text-red-700 px-1 py-0.5 border-2 border-rose-600 bg-rose-400 scale-96 group-hover:bg-orange-50 group-hover:opacity-75 group-hover:scale-100 hover:opacity-100 hover:scale-96 hover:bg-rose-400 transition-all"
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
          <div className="select-none group">
            <span className="flex justify-between items-start">
              <p className="item-title">{item.Title}</p>
              <button
                className="rounded-[2vw] text-sm bg-slate-100 border-slate-300 border-2 px-1 text-slate-500 hover:bg-slate-200 hover:scale-96 transition-all group-hover:scale-96 group-hover:bg-slate-200"
                onClick={() => setEditingId(item.ID)}
              >
                Edit
              </button>
            </span>
            <p className="item-body overflow-wrap">{item.Body}</p>
            <p className="item-timestamp">Created - {formatDate(item.CreatedAt).toLocaleDateString()}</p>
          </div>
        )}
      </li>
    </>
  );
});

export default Item;
