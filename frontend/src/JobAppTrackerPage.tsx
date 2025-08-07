import { useState, useEffect } from 'react'
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, updateTrackerItem, formatDate } from './api/tracker'
import { useTrackerData } from './JobAppTrackerDataContext';
import { useScreenSize } from './ScreenSizeProvider';
import './App.css'

function formatMinutes(minLeft:number) {
  let d = Math.floor(minLeft/60/24)
  let h = Math.floor(minLeft/60 - d * 24);
  let m = (minLeft % 60);
  console.log(d, h, m)
  return ((d > 0)? d.toString() + "d ": "") + ((h > 0)? h.toString() + "h ": "") + ((m > 0)? m.toString() + "m ": "");
}

function JobAppTrackerPage() {
  const isDesktop = useScreenSize();
  const { user, tracker, error } = useTrackerData();  
  if (error) return <div>Error loading backend: {error}</div>;
  if ( !user || !tracker ) return <div>???</div>;

  return (
    <div className="px-8 pt-4 w-8/10 justify-self-center border-blue-200 border mt-3">
      <h1 className="vp-mid text-5xl font-bold select-none text-indigo-700 mb-4 text-center text-shadow-2xs text-shadow-blue-300">
        {isDesktop? 'Job App Tracker With Lootbox Technology + Agentic Functionality' : 'Job App Tracker'}
      </h1>
      <TrackerBar />
      <ItemsList />
    </div>
  );
}

function ProgressBoxes() {
  const { tracker, error, setTracker } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;

  const gq = tracker.GoalQuantity;
  const completed = tracker.CurScorableItems % gq;
  const blocks = [];

  for (let i = 0; i < gq; i++) {
    console.log(i, completed)
    blocks.push(
    <li key={i} className={`min-h-2 min-w-3 flex-1 mx-1.5 rounded transition-colors duration-200 ${(i < completed)? "bg-lime-500" : "bg-slate-700"}`}> </li>);
  };

  return (
    <ul className='flex items-center justify-between border min-h-2.5'>{blocks}</ul>
  );
}

function ProgressTracker() {
  const { tracker, error } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  
  // local timer:
  // https://medium.com/create-a-clocking-in-system-on-react/create-a-react-app-displaying-the-current-date-and-time-using-hooks-21d946971556
  const [date, setDate] = useState(new Date());
  useEffect(() => {
    // update the date every min
    const interval = 60 * 1000
    var timer = setInterval(() => setDate(new Date()), interval)
    return function cleanup() {
      clearInterval(timer)
    }
  })

  const deadline = formatDate(tracker.CycleDeadline)
  let deadlineMinutes = 60 * 24
  if (tracker.CycleFrequency == 'weekly') {
    deadlineMinutes = 60 * 24 * 7
  }

  let minsRemaining = Math.floor((deadline.valueOf() - date.valueOf())/ 1000 / 60)
  return (
    <span className='px-2 pt-0.5 pb-2'>
      <p className='text-xs text-left'> Remaining applications until next lootbox: {tracker.GoalQuantity - tracker.CurScorableItems} </p>
      <ProgressBoxes />
      <label htmlFor="remTime" className='text-xs'> Time remaining: {formatMinutes(minsRemaining)}</label>
      <progress id="remTime" value={minsRemaining} max={deadlineMinutes}></progress>
    </span>
  );
}

function TrackerBar() {
  const { tracker, error, setTracker } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  const deadline = formatDate(tracker.CycleDeadline).toISOString()

  const [isEditing, setIsEditing] = useState(false);
  return (
      
      <div className="min-h-[79px] bg-blue-400 border-x-violet-300 border-4 ">
        <div className='flex justify-between px-1.5 mb-1 text-center'>
          <ProgressTracker />
          <p>Current lootboxes: {tracker.CurBoxesAwarded}</p>
          <button className='text-xs align-right border-2 rounded-[2vw] max-h-6 mt-4' onClick={() => setIsEditing(!isEditing)}> settings </button>

        </div>
        {isEditing && <span>
          <div> Set the target amount of applications to unlock a lootbox: </div>
          <input type="number" defaultValue={tracker.GoalQuantity}></input>
          <div> Set the deadline to reset your progress to a lootbox: </div>
          
          <input type='datetime-local' value={deadline.substring(0, deadline.indexOf('T')+6)}></input>
          {/* https://developer.mozilla.org/en-US/docs/Web/HTML/Guides/Date_and_time_formats#local_date_and_time_strings */}
          <div>{formatDate(tracker.CycleDeadline).toJSON()}</div>
          <div>{formatDate(tracker.CycleDeadline).toLocaleString()}</div>
          <div>{formatDate(tracker.CycleDeadline).toLocaleDateString()}</div>
          <div>{formatDate(tracker.CycleDeadline).toLocaleTimeString()}</div>
          <div>{formatDate(tracker.CycleDeadline).toISOString()}</div>
          <div>GoalQuantity</div>
          
        </span>}
      </div>

    
  )
}

function ItemsList() {
  const isDesktop = useScreenSize();
  const { user, tracker, error, refreshData, setTracker } = useTrackerData();
  const [isNewItem, setIsNewItem] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);

  if (error) return <div>Error loading backend: {error}</div>;

  const blankItem = { Title: "", Body: "" };

  const handleNew = async (newTitle: string, newBody: string) => {
    try {
      const newData = await createTrackerItem(newTitle, newBody);
      newData.tracker ? setTracker(newData.tracker) : console.error("Error finding data from Backend");
    } catch (error: any) {
      console.error(error.message);
    }
  };

  const handleEdit = async (item: any, newTitle: string, newBody: string) => {
    if (item.Title === newTitle && item.Body === newBody) return;
    try {
      const newData = await updateTrackerItem(item.ID, newTitle, newBody);
      newData.tracker ? setTracker(newData.tracker) : console.error("Error finding data from Backend");
    } catch (error: any) {
      console.error(error.message);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteTrackerItem(id);
    } catch (error: any) {
      console.error(error.message);
    }
    refreshData();
  };

  return (
    <div className='justify-self-center border-2 w-full max-w-170'>
      {isNewItem ? (
        <Item
          item={blankItem}
          isEditing={editingId === "new"}
          isNewItem={true}
          setIsNewItem={setIsNewItem}
          editingId={editingId}
        setEditingId={setEditingId}
          handleEdit={handleEdit}
          handleDelete={handleDelete}
          handleNew={handleNew}
        />
      ) : (
        <button
          className="flex rounded-4xl mt-2 mb-1 py-1 px-3 border-2 bg-blue-300 justify-self-center text-md hover:bg-blue-400"
          onClick={() => {
            setIsNewItem(true);
            setEditingId("new");
          }}
        >
          {isDesktop ? "new application" : "+"}
        </button>
      )}
      <ul className="space-y-2">
        {tracker?.Items.map((item: any) => (
          <Item
            key={item.ID}
            item={item}
            isEditing={editingId === item.ID}
            isNewItem={false}
            setIsNewItem={setIsNewItem}
            editingId={editingId}
            setEditingId={setEditingId}
            handleEdit={handleEdit}
            handleDelete={handleDelete}
            handleNew={handleNew}
          />
        ))}
      </ul>
      <p>user: {user?.ID}</p>
    </div>
  );
}



export default JobAppTrackerPage
