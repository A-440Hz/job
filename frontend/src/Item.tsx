import React, { useEffect, useState } from 'react';
import { formatDate } from './api/datetime';
import { fetchScraperData, summarizeScrapedData } from './api/scraper';
import { useTrackerData } from './JobAppTrackerDataContext';

// https://tw-elements.com/docs/standard/components/spinners/
const spinner = (
  <div
    className="inline-block w-6 aspect-square mr-1 animate-spin rounded-full border-4 border-solid border-orange-400 border-e-transparent align-[-0.125em] text-surface motion-reduce:animate-[spin_1.5s_linear_infinite] dark:text-white"
    role="status">
    <span
      className="!absolute !-m-px !h-px !w-px !overflow-hidden !whitespace-nowrap !border-0 !p-0 ![clip:rect(0,0,0,0)]">
        Loading...
    </span>
  </div>
);

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
  handleEdit: (item: any, title: string, body: string, url: string) => Promise<boolean>;
  handleDelete: (id: string) => void;
  handleNew: (title: string, body: string, url: string) => Promise<boolean>;
}) {
  const [title, setTitle] = useState(item.Title);
  const [url, setUrl] = useState(item.Url);
  const [body, setBody] = useState(item.Body);
  const [saveHighlight, setSaveHighlight] = useState(true);
  const [queryState, setQueryState] = useState<'noQuery' | 'scraping' | 'processing' | 'processed'>('noQuery');
  const {modelName} = useTrackerData();

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

  useEffect(() => {
    const autoSubmit = async () => {
      if (queryState === 'processed') {
        await submitChanges()
        setQueryState('noQuery');
      }
    }
    autoSubmit();
  }, [queryState]);

  const exitEditing = () => {
    setEditingId(null);
    if (item.ID === undefined) {
      setIsNewItem(false);
    }
  };

  const requestSummary = async (content: string, itemId: string) => {
    summarizeScrapedData(content, itemId, modelName)
      .then(data => {
        if (data && data.summary) { // TODO: check for non-cancelled state before setting
          console.log(data);
          setBody(data.summary);
          setTitle(data.position_name + ' - ' + data.company_name);
          setQueryState('processed');
        }
      })
      .catch(error => {
        console.error('Error summarizing scraped data:', error);
        setQueryState('noQuery');
    });
  };

  const submitURL = () => {
    if (queryState !== 'noQuery') return;
    if (!url && body === '') return;

    // send directly to processing if no url provided
    if (!url && body !== '') {
      setQueryState('processing');
      let content = body;
      if (title !== '') {
        content = title + "\n\n" + content;
      }
      requestSummary(content, item.ID);
      return;
    }

    // send to scraper if url exists
    setQueryState('scraping');
    fetchScraperData(url, item.ID)
      .then((data) => {
        if (data && (data.code !== 200 || data.content === "")) {
          console.error('Error fetching scraper data:', data.error ? data.error : 'Unknown error');
          console.log(data.content === "" && 'Failed to scrape anything from the url.');
          // console.error('url refused to be scraped.');
          if (body === '') {
            console.log('Try pasting the job description into notes and pressing "AI" again to directly process it.');
          } else {
            console.log('Sending current notes to LLM to generate a summary...');
            setQueryState('processing');
            requestSummary(body, item.ID);
            return;
          }
          setQueryState('noQuery');
          return;
        }
        if (data && data.content) {
          setBody(data.content);
          setQueryState('processing');
          requestSummary(data.content, item.ID);
          return;
        }
      })
      .catch((err) => {
        console.error('Error fetching scraper data:', err);
        setQueryState('noQuery');
        // make the url box flash red
        // error msg "failed to scrape data from URL"
      });
  }

  const submitChanges = async () => {
    if (item.ID === undefined) {
      if (!title) return; // Prevent blank titles. TODO:a flashing animation for the title input border 
      let success = await handleNew(title, body, url);
      if (success === true && (queryState === 'scraping' || queryState === 'processing')) {
        setQueryState('noQuery');
      }
    } else {
      let success = await handleEdit(item, title, body, url);
      if (success === true && (queryState === 'scraping' || queryState === 'processing')) {
        setQueryState('noQuery');
      }
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
            <span className={`flex items-start justify-between ${queryState !== 'noQuery' && "select-none pointer-events-none"}`} >
              {queryState === 'scraping' || queryState === 'processing' ? spinner : (
                <div className="flex w-7 justify-center bg-slate-200 text-slate-700 border aspect-square select-none transition hover:bg-slate-300 hover:scale-79"
                  onClick={(e) => {
                    e.stopPropagation();
                    submitURL();
                  }
                }
                ><strong>AI</strong></div>
              )}
              <input
                  className="text-amber-700 input-box"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder=" url"
                  onMouseOver={() => setSaveHighlight(false)}
                  onMouseLeave={() => setSaveHighlight(true)}
              />
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
              <span className="flex justify-between items-start">
                {queryState === 'scraping' || queryState === 'processing' ? spinner : null}
                <p className="item-title">{item.Title}</p>
              </span>
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
