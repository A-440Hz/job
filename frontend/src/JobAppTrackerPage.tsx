import { useRef, useState, useEffect } from 'react'
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, updateTrackerItem, updateTracker } from './api/tracker'
import { formatDate, dateToInputString, inputStringToDate, adjustTimezoneOffset, convertToBackendTime } from './api/datetime'
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

function JobAppTrackerPage() {
  const { user, tracker, error } = useTrackerData();
  const isDesktop = useScreenSize();
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
  const { tracker, error } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;

  const gq = tracker.GoalQuantity;
  const completed = tracker.CurScorableItems % gq;
  const blocks = [];

  for (let i = 0; i < gq; i++) {
    // console.log(i, completed)
    blocks.push(
    <li key={i} className={`min-h-2 min-w-3 max-w-13 flex-1 mx-1.5 rounded transition-colors duration-200 ${(i < completed)? "bg-lime-500 opacity-80 shadow-xs shadow-amber-50" : "bg-slate-700"}`}> </li>);
  };

  return (<div className='select-none'>
    <div className='bg-lime-600 w-20 skew-[-6deg] flex my-1'>
      <span className='text-left font-semibold indent-4 text-xl block skew-[6deg]'> Goal: </span>
    </div>
    <div className='text-xs text-left text-nowrap'> 
      Complete 
        <span className='bg-lime-600 inline-flex px-1 mx-0.5 skew-[6deg]'>
          <span className='skew-[-6deg] font-semibold'> {tracker.GoalQuantity} </span>
        </span>
      application<span>{tracker.GoalQuantity > 1 && 's'}</span> to earn a lootbox
    </div>
    <ul className='flex mt-2.5 items-center justify-between min-h-2.5'>{blocks}</ul>
  </div>
  );
}

function DeadlineBar() {
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

  return (<div className='select-none flex-col justify-items-left'>
  <div className='bg-amber-900 w-28 skew-[-3deg] flex my-1'>
      <span className='text-left font-semibold indent-4 text-xl block skew-[3deg]'> Deadline: </span>
    </div>
    <label htmlFor="remTime" className='text-xs flex text-left text-nowrap'> Progress resets in: 
      <span className='bg-amber-900 inline-flex px-1 mx-0.5 skew-[3deg]'>
        <span className='skew-[-3deg] font-semibold'> {formatMinutes(minsRemaining)} </span>
      </span>      
    </label>
    <progress id="remTime" className='mt-3 flex w-[100%]' value={minsRemaining} max={deadlineMinutes}></progress>
  </div>)
}

function ProgressTracker() {
  const isDesktop = useScreenSize()
  return (
    (isDesktop)?
      <>
    <span className='px-2 pt-0.5 pb-2'>
      <ProgressBoxes />
    </span>
    <span className='px-2 pt-0.5 pb-2 align-self-center block'>
      <DeadlineBar />
    </span>
      </>:
    <span className='px-2 pt-0.5 pb-2'>
      <ProgressBoxes />
      <DeadlineBar />
    </span>
    
  );
}

function TrackerBar() {
  const { tracker, user, error, setTracker } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  
  const [isEditing, setIsEditing] = useState(false);
  const [saveHighlight, setSaveHighlight] = useState(true);

  useEffect(() => {
    if (!isEditing) {
      setDeadline(localizedDeadline);
      setQuantity(tracker.GoalQuantity);
      setFrequency(tracker.CycleFrequency);
      setPenalty(tracker.MissedGoalPenalty);
    }
  })

  const now = new Date();
  const localizedDeadline = dateToInputString(adjustTimezoneOffset(formatDate(tracker.CycleDeadline), user.Timezone.OffsetSeconds, true));
  const [deadline, setDeadline] = useState(localizedDeadline);
  const [quantity, setQuantity] = useState(Number(tracker.GoalQuantity));
  const [frequency, setFrequency] = useState(tracker.CycleFrequency);
  const [penalty, setPenalty] = useState(Boolean(tracker.MissedGoalPenalty));

  const exitEditing = () => {
    setIsEditing(false);
  }

  const submitChanges = () => {
    const newDeadline = deadline;
    const newQuantity = quantity;
    const newFrequency = frequency;
    const newPenalty = penalty;
    // console.log(deadline)
    console.log(newPenalty,  newFrequency, newDeadline, newQuantity)
    
    handleEdit(
      newPenalty,
      newFrequency,
      newDeadline,
      newQuantity,
    );
    exitEditing();
  }

  const handleEdit = async (penalty:boolean, frequency?:string, deadline?:string, quantity?:number) => {
    const changes: {
      penalty?: boolean;
      frequency?: string;
      deadline?: number;
      quantity?: number;
    } = {};

    if (frequency && frequency !== tracker.CycleFrequency) {changes.frequency = frequency;}
    // convert microseconds to seconds for backend
    if (deadline !== localizedDeadline) {changes.deadline = convertToBackendTime(adjustTimezoneOffset(inputStringToDate(String(deadline)), user.Timezone.OffsetSeconds, false));}
    if (quantity && quantity !== tracker.GoalQuantity) {changes.quantity = quantity;}
    if (penalty !== tracker.MissedGoalPenalty) {changes.penalty = penalty;}

    if (Object.keys(changes).length === 0) return;

    try {
      console.log(changes)
      const newData = await updateTracker(
        changes.frequency,
        changes.deadline,
        changes.quantity,
        changes.penalty
      );
      console.log(newData)
      newData.tracker ? setTracker(newData.tracker) : console.error("Error finding data from Backend");
    } catch (error: any) {
      console.error(error.message);
    }
  };
  
  return (      
      <div className="min-h-[79px] bg-blue-400 border-x-violet-300 border-4 ">
        <div className='flex justify-between px-1.5 mb-1 text-center'>
          <ProgressTracker />
          <p className='text-sm p-1'>Current lootboxes earned {(tracker.CycleFrequency === "weekly")? 'this week' : 'today'}: {tracker.CurBoxesAwarded}</p>
          <button className='text-xs align-right border-2 rounded-[2vw] max-h-6 mt-4' onClick={() => {isEditing? exitEditing(): setIsEditing(true)}}> settings </button>
        </div>
        {isEditing && <div className='flex-row bg-gray-500 opacity-70'>
        <div className='grid grid-cols-2 p-2 gap-4'>
          <div className={`border-3 p-2 flex-col justify-between ${(quantity != tracker.GoalQuantity) && 'border-amber-300'}`}>
            <div className='block'> Set the goal for earning a lootbox: </div>
            <input className='bg-lime-500 w-10 ml-1 block justify-self-center' type="number" value={quantity} min={1} onChange={(e) => setQuantity(e.target.valueAsNumber)}></input>
          </div>
          <div className={`border-3 p-2 ${(deadline !== localizedDeadline) && 'border-amber-300'}`}>
            <label htmlFor="select-date"> Set goal deadline: </label>
            <input className='border' id='select-date' type='datetime-local' value={deadline} min={dateToInputString(now)} onChange={(e) => setDeadline(e.target.value)}></input>
            <p> {Math.floor(formatDate(tracker.CycleDeadline).valueOf()/ 1)} </p>
            <p> {Math.floor(adjustTimezoneOffset(inputStringToDate(deadline), user.Timezone.OffsetSeconds, false).valueOf()/1)} </p>
            <p> {adjustTimezoneOffset(inputStringToDate(deadline), user.Timezone.OffsetSeconds, false).toJSON()} </p>
            <p> {formatDate(tracker.CycleDeadline).toJSON()} </p>
          </div>
          {/* <input type='datetime-local' value={addOffsetSeconds(tracker.CycleDeadline, user.Timezone.Offset).toISOString().substring(0, deadline.indexOf('T')+6)} ref={deadlineRef}></input> */}
          {/* https://developer.mozilla.org/en-US/docs/Web/HTML/Guides/Date_and_time_formats#local_date_and_time_strings */}
          {/* <div>{deadlineRef.current.value};</div> */}
          <div className={`border-3 p-2 ${(frequency !== tracker.CycleFrequency) && 'border-amber-300'}`}> 
            <div> Set deadline reset frequency: </div>
            <select value={frequency} onChange={(e) => setFrequency(e.target.value)}>
              <option className='text-gray-300' value="daily">daily</option>
              <option className='text-gray-300' value="weekly">weekly</option>
            </select>
          </div>
          <div className={`border-3 p-2 ${(penalty !== tracker.MissedGoalPenalty) && 'border-amber-300'}`}>
            <div> Enable goal failure penalty:
            <input className='ml-1' type="checkbox" defaultChecked={penalty} onChange={() => setPenalty(!penalty)}></input>
            </div>
          </div>
        </div>
        <span className="float-right border-1">
          <button
            className={`mr-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 ${
              saveHighlight ? "hover:bg-emerald-500 hover:opacity-35" : ""
            }`}
            onClick={(e) => {
              e.stopPropagation();
              submitChanges();
            }}
          >
            Save
          </button>
          <button
            className="ml-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 bg-rose-400 opacity-35 hover:bg-blue-400 hover:opacity-100 hover:bg-rose-400 hover:opacity-35"
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
        </div>}
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
        {tracker?.Items?.map((item: any) => (
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
