import { useState, useEffect } from 'react'
import { createTrackerItem, updateTrackerItem } from './api/tracker'
import { useTrackerData } from './JobAppTrackerDataContext';
import { useScreenSize } from './ScreenSizeProvider';
import './App.css'


function formatDate(raw: string) {
  const d = new Date(raw);
  return d.toLocaleDateString();
}

function JobAppTrackerPage() {
  const isDesktop = useScreenSize();
  const { user, tracker, error } = useTrackerData();  
  if (error) return <div>Error loading backend: {error}</div>;
  if ( !user || !tracker ) return <div>???</div>;

  return (
    <div className=" px-8 pt-4 w-8/10 justify-self-center border-blue-200 border mt-3">
      {/* TODO: use CSS to refactor out isDesktop effect */}
      <h1 className="vp-mid text-5xl font-bold select-none text-indigo-700 mb-4 text-center text-shadow-2xs text-shadow-blue-300">
        {isDesktop? 'Job App Tracker With Lootbox Technology + Agentic Functionality' : 'Job App Tracker'}
      </h1>
      <div className="text-xl mb-2 bg-blue-400 border-x-violet-300 border-4 text-center">maybe this is a topbar for options and sort order</div>
      <ItemsList />
    </div>
  );
}

function ItemsList() {
  const isDesktop = useScreenSize();
  const { user, tracker, error, setTracker } = useTrackerData();
  const [ isNewItem, setIsNewItem ] = useState(false);
  if (error) return <div>Error loading backend: {error}</div>;

  // type blankItem = ReturnType<typeof item>;
    const blankItem = {
      Title: '',
      Body: '',
    }

  const handleNew = async(newTitle: string, newBody: string) => {
    try {
      const newData = await createTrackerItem(newTitle, newBody);
      // TODO: dont have this;
      console.log(newData);
      console.log("tracker:", newData.tracker);
      console.log("t:", newData.t);
      (newData.tracker? setTracker(newData.tracker) : null);
      (newData.t? setTracker(newData.t) : null);
    } catch (error: any) {
      console.log(error.message);
    }
  };

  const handleEdit = async (item: any, newTitle: string, newBody: string) => {
    if (item.Title == newTitle && item.Body == newBody) {
      return
    }
    try {
      const newData = await updateTrackerItem(item.ID, newTitle, newBody);
      // TODO: dont have this;
      
      console.log(newData);
      console.log("tracker:", newData.tracker);
      console.log("t:", newData.t);
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

    useEffect(() => {
      if (item.ID === undefined) {
        setIsEditing(true);
      }
    }, [isEditing])

    function exitEditing() {
      setIsEditing(false);
      if (item.ID === undefined) {
        setIsNewItem(false);
        return;
      }
      setTitle(item.Title);
      setBody(item.Body);
    }

    function submitChanges(item: any, title: string, body: string) {
      if (item.ID === undefined) {
        // console.log(title, body)
        return submitNew(title, body)
      }
      handleEdit(item, title, body);
      setIsEditing(false);
    }

    function submitNew(title: string, body: string) {
      if (title === undefined) {
        // flash red and don't submit
        return;
      }
      handleNew(title, body);
      setIsNewItem(false);
      setIsEditing(false);
    }


    // ideally i overload the submitChanges and exitEditing functions depending on if the item
    // is new or not. Just 3 if cases in the return and 
    return (
      <>
        {isEditing && (
          <div className='item-modal' onClick={() => exitEditing()}/>
        )}
      <li key={item.ID} className={`py-4 pl-6 pr-9 rounded bg-orange-200 shadow relative ${isEditing ? 'item-editing': ''}`} onClick={() => { isEditing? submitChanges(item, title, body): setIsEditing(true) }}>
        {isEditing ? (
          <>
          <div onClick={e => e.stopPropagation()}>
            <input className="item-title input-box " value={title} onChange={e => setTitle(e.target.value)} placeholder={"*Company - Position"}/>
            <textarea className="item-body input-box " value={body} rows={5} onChange={e => setBody(e.target.value)} placeholder="notes" />
          </div>
            <p className={`text-xs flex text-gray-900 ${isNewItem? 'hidden' : ''}`}> Created - {formatDate(item.CreatedAt)} </p>
            <span className="float-right flex m-2">
              <button className="mr-2" onClick={(e) => {e.stopPropagation(); submitChanges(item, title, body) }}>Save</button>
              <button className="ml-2" onClick={(e) => {e.stopPropagation(); exitEditing() }}>Cancel</button>
            </span>
          </>
        ) : (
          <>
          <div className='select-none' >
            <span className='flex justify-between items-center'>
              <p className="item-title">{item.Title}</p>
              {/* <svg className="hover:" onClick={() => setIsEditing(true)} width="64px" height="64px" viewBox="-61.75 -61.75 185.25 185.25" xmlns="http://www.w3.org/2000/svg" fill="#000000" stroke="#000000" stroke-width="0.494008"><g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round" stroke="#CCCCCC" stroke-width="0.7410119999999999"></g><g id="SVGRepo_iconCarrier"> <path id="Path_3" data-name="Path 3" d="M328.667,219.141l-7.779-7.838a5.935,5.935,0,0,0-4.226-1.746h0a5.933,5.933,0,0,0-4.225,1.745l-38.183,38.174a1.512,1.512,0,0,0-.391.683l-5.015,19.271a1.5,1.5,0,0,0,1.452,1.878,1.472,1.472,0,0,0,.388-.051l19.154-5.132a1.49,1.49,0,0,0,.673-.389l38.148-38.147a5.989,5.989,0,0,0,.005-8.448ZM307.4,220.574l4.928,4.928-29.966,29.966-4.931-4.931Zm-31.465,46.17-2.613-2.613,2.78-10.681,10.45,10.449Zm13.514-4.189-4.966-4.966,29.966-29.966,4.966,4.966Zm37.088-37.088-5,5-12.014-12.015,5.031-5.03h0a2.953,2.953,0,0,1,2.1-.866h0a2.951,2.951,0,0,1,2.1.863l7.78,7.838a2.989,2.989,0,0,1,0,4.209Z" transform="translate(-268.799 -209.557)" fill="#5b4933"></path> </g></svg> */}
              <button className="rounded-[2vw] text-sm bg-blue-100 border-blue-300 border-2 px-1 text-amber-800 hover:bg-blue-200" onClick={() => setIsEditing(true)}>Edit</button>
            </span>
            <p className="item-body">{item.Body}</p>
            <p className='item-timestamp'> Created - {formatDate(item.CreatedAt)} </p>
          </div>
          </>
        )}
      </li>
    </>
    );
  }

  return (
    
    <>
    {/* <p className='text-blue-300'> HIHIHIHIH {isNewItem == true? 'treu' : 'nop'} </p> */}
    {isNewItem? <Item item={blankItem}></Item> : <button 
      className={`flex rounded-4xl py-1 px-3 mb-0.5 border-2 bg-blue-300 justify-self-center hover:bg-blue-400`}
      onClick={() => setIsNewItem(true)}>
      {isDesktop? 'new application' : '+'}
    </button>}
    <ul className='space-y-2'>
      {tracker?.Items.map((item: any) => (
        <Item key={item.ID} item={item} />
      ))}
    </ul>
    <p> user: {user?.ID}</p>
    </>
  );
}


export default JobAppTrackerPage
