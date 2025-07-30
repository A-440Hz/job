import { useEffect, useState } from 'react'
import { updateTrackerItem } from './api/tracker'
import { useTrackerData } from './JobAppTrackerDataContext';
import './App.css'


function formatDate(raw: string) {
  const d = new Date(raw);
  return d.toLocaleDateString();
}

function JobAppTrackerPage() {
  const { user, tracker, error } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  if ( !user || !tracker ) return <div>???</div>;

  return (
    <div className="p-8 w-8/10 justify-self-center border-blue-200 border mt-8">
      <h1 className="text-5xl font-bold text-indigo-700 mb-4 text-center ">
        Job App Tracker With Lootbox Technology
      </h1>
      <div className="text-xl mb-2 bg-violet-500 border-x-violet-500 border-4 text-center">Your Job Applications:</div>
      <ItemsList />
    </div>
  );
}

function ItemsList() {
  const { user, tracker, error, setTracker } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  // const [items, setItems] = useState(data.tracker?.Items ?? []);

  const handleEdit = async (id: string, newTitle: string, newBody: string) => {
    try {
      const newData = await updateTrackerItem(id, newTitle, newBody);
      // TODO: dont have this;
      (newData.tracker? setTracker(newData.tracker) : null);
      (newData.t? setTracker(newData.t) : null);
      // TODO: error handling
      ;
    } catch (error: any) {
        console.log(error.message);
    } 
  };  

  function Item({ item }: { item: any }) {
    const [isEditing, setIsEditing] = useState(false);
    const [title, setTitle] = useState(item.Title);
    const [body, setBody] = useState(item.Body);

    return (
      <li key={item.ID} className={`p-4 rounded bg-orange-200 shadow ${isEditing ? 'item-editing': ''}`}>
        {isEditing ? (
          <>
            <input className="item-title input-box " value={title} onChange={e => setTitle(e.target.value)} />
            <br></br>
            <textarea className="item-body input-box " value={body} rows={5} onChange={e => setBody(e.target.value)} />
            <p className='text-xs text-gray-900'> Created - {formatDate(item.CreatedAt)} </p>

            <button onClick={() => { handleEdit(item.ID, title, body); setIsEditing(false); }}>Save</button>
            <button onClick={() => {setIsEditing(false); setTitle(item.Title); setBody(item.Body)}}>Cancel</button>
          </>
        ) : (
          <>
            <p className="item-title">{item.Title}</p>
            <p className="item-body">{item.Body}</p>
            <p className='item-timestamp'> Created - {formatDate(item.CreatedAt)} </p>
            <button className="rounded-[2vw] bg-amber-500" onClick={() => setIsEditing(true)}>Edit</button>
          </>
        )}
      </li>
    );
  }

  return (
    <>
    <ul className='space-y-2'>
      {tracker?.Items.map((item: any) => (
        <Item item={item} />
      ))}
    </ul>
    <p> user: {user?.ID}</p>
    </>
  );
}


export default JobAppTrackerPage
