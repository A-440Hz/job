import { useRef, useState, useEffect } from 'react'
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, updateTrackerItem, updateTracker, formatDate } from './api/tracker'
import { useTrackerData } from './JobAppTrackerDataContext';
import { useScreenSize } from './ScreenSizeProvider';
import './App.css'

function formatMinutes(minLeft:number) {
  let d = Math.floor(minLeft/60/24)
  let h = Math.floor(minLeft/60 - d * 24);
  let m = (minLeft % 60);
  // console.log(d, h, m)
  return ((d > 0)? d.toString() + "d ": "") + ((h > 0)? h.toString() + "h ": "") + ((m > 0)? m.toString() + "m ": "");
}

function addOffsetSeconds(date:Date, seconds:number) {
  return new Date(date.valueOf()+seconds)
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
    // console.log(i, completed)
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
  const { tracker, user, error, setTracker } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  const deadline = formatDate(tracker.CycleDeadline).toISOString()

  const [isEditing, setIsEditing] = useState(false);
  const [saveHighlight, setSaveHighlight] = useState(true);

  const deadlineRef = useRef<HTMLInputElement>(tracker.CycleDeadline)
  const frequencyRef = useRef<HTMLSelectElement>(tracker.CycleFrequency)
  const quantityRef = useRef<HTMLInputElement>(tracker.GoalQuantity)
  const penaltyRef = useRef<HTMLInputElement>(tracker.MissedGoalPenalty)

  const submitChanges = () => {
    const newDeadline = deadlineRef.current.value;
    const newQuantity = Number(quantityRef.current.value);
    const newFrequency = frequencyRef.current.value;
    const newPenalty = Boolean(penaltyRef.current.value);
    console.log(newPenalty)
    handleEdit(
      newFrequency,
      newQuantity,
      newPenalty,
    );
    setIsEditing(false);
  }

  const handleEdit = async (frequency?:string, quantity?:number, penalty?:boolean) => {
    if (tracker.CycleDeadline === deadline && tracker.CycleFrequency === frequency && tracker.GoalQuantity === quantity && tracker.MissedGoalPenalty === penalty) return;
    try {
      const newData = await updateTracker(frequency, quantity, penalty);
      newData.tracker ? setTracker(newData.tracker) : console.error("Error finding data from Backend");
    } catch (error: any) {
      console.error(error.message);
    }
  };

  return (      
      <div className="min-h-[79px] bg-blue-400 border-x-violet-300 border-4 ">
        <div className='flex justify-between px-1.5 mb-1 text-center'>
          <ProgressTracker />
          <p>Current lootboxes: {tracker.CurBoxesAwarded}</p>
          <button className='text-xs align-right border-2 rounded-[2vw] max-h-6 mt-4' onClick={() => setIsEditing(!isEditing)}> settings </button>

        </div>
        {/* look into <form> https://react.dev/reference/react-dom/components/input*/}
        {isEditing && <span>
          <div> Set the target amount of applications to unlock a lootbox: </div>
          <input type="number" defaultValue={tracker.GoalQuantity} ref={quantityRef}></input>
          <div> Set the deadline to reset your progress to a lootbox: </div>
          <input type='datetime-local' value={deadline.substring(0, deadline.indexOf('T')+6)} ref={deadlineRef}></input>
          {/* <input type='datetime-local' value={addOffsetSeconds(tracker.CycleDeadline, user.Timezone.Offset).toISOString().substring(0, deadline.indexOf('T')+6)} ref={deadlineRef}></input> */}
          {/* https://developer.mozilla.org/en-US/docs/Web/HTML/Guides/Date_and_time_formats#local_date_and_time_strings */}
          <div>{formatDate(tracker.CycleDeadline).toJSON()}</div>
          <div>{formatDate(tracker.CycleDeadline).toLocaleString()}</div>
          <div>{formatDate(tracker.CycleDeadline).toLocaleDateString()}</div>
          <div>{formatDate(tracker.CycleDeadline).toLocaleTimeString()}</div>
          <div>{formatDate(tracker.CycleDeadline).toISOString()}</div>
          <div> Set the frequency at which your goal resets: </div>
          <select ref={frequencyRef} defaultValue={tracker.CycleFrequency}>
            <option value="weekly">weekly</option>
            <option value="daily">daily</option>
          </select>
          <input type="checkbox" ref={penaltyRef} defaultChecked={tracker.MissedGoalPenalty}></input>
          <span className="float-right flex border-1 border-amber-500">
                <button
                  className={`mr-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 ${
                    saveHighlight ? "group-hover:bg-emerald-500 group-hover:opacity-35" : ""
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
                    setIsEditing(false);
                  }}
                  onMouseOver={() => setSaveHighlight(false)}
                  onMouseLeave={() => setSaveHighlight(true)}
                >
                  Cancel
                </button>
              </span>
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
        <div className='flex justify-center'>
        <button
          className="rounded-4xl select-none mt-2 mb-1 py-1 px-3 border-2 bg-blue-300 justify-self-center text-md hover:bg-blue-400"
          onClick={() => {
            setIsNewItem(true);
            setEditingId("new");
          }}
        >
          {isDesktop ? "new application" : "+"}
        </button>
        </div>
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
