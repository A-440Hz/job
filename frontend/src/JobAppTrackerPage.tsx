import { useRef, useState, useEffect, type ReactNode } from 'react'
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, updateTrackerItem, updateTracker } from './api/tracker'
import { formatDate, dateToInputString, inputStringToDate, adjustTimezoneOffset, convertToBackendTime } from './api/datetime'
import { useTrackerData } from './JobAppTrackerDataContext';
import { useScreenSize } from './ScreenSizeProvider';
import flameHot from './assets/flame-hot-svgrepo-com.svg';
import flameCold from './assets/flame-cold-svgrepo-com.svg';
import gearIcon from './assets/gear-svgrepo-com.svg';

import './App.css'

function formatMinutes(minLeft:number) {
  if (minLeft < 0) {return -1}
  let d = Math.floor(minLeft/60/24)
  let h = Math.floor(minLeft/60 - d * 24);
  let m = (minLeft % 60);
  // console.log(d, h, m)
  return ((d > 0)? d.toString() + "d ": "") + ((h > 0)? h.toString() + "h ": "") + ((m > 0)? m.toString() + "m ": "");
}

function JobAppTrackerPage() {
  const { user, tracker, error, refreshData } = useTrackerData();
  const isDesktop = useScreenSize();

  useEffect(() => {
    if (user) {
      refreshData();
    }
  }, [user, refreshData]);

  if (error) return <div>Error loading backend: {error}</div>;
  if (!user || !tracker) return <div>???</div>;

  return (
    <div className="px-8 pt-4 w-8/10 justify-self-center border-blue-200 border mt-3">
      <h1 className="vp-mid text-5xl font-bold select-none text-indigo-700 mb-4 text-center text-shadow-2xs text-shadow-blue-300">
        {isDesktop? 'Job App Tracker With Lootbox Technology + Agentic Functionality' : 'Job App Tracker'}
      </h1>

      <div className={`${isDesktop?'':'w-[90vw] relative left-1/2 right-1/2 -mx-[45vw]'} `}>
        <TrackerBar />
      </div>
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
    blocks.push(
    <li key={i} className={`h-3 min-w-4 flex-1 mx-1 rounded-md transition-all duration-300 ease-in-out ${(i < completed)? "bg-emerald-500 shadow-sm scale-105" : "bg-gray-300 hover:bg-gray-400"}`}> </li>);
  };

  return (<div className='bg-white rounded-xl p-4 shadow-md border border-gray-100 min-w-0 flex-1'>
    <div className='flex items-center justify-between mb-3'>
      <h3 className='text-lg font-semibold text-gray-800 flex items-center gap-2'>
        <span className='w-2 h-2 bg-emerald-500 rounded-full'></span>
        Goal Progress
      </h3>
      <span className='text-xs font-medium text-gray-500 bg-gray-50 px-2 py-1 rounded-full'>
        {tracker.CurBoxesAwarded} earned {(tracker.CycleFrequency === "weekly")? 'this week' : 'today'}
      </span>
    </div>
    <div className='text-sm text-gray-600 mb-3'>
      Complete <span className='font-semibold text-emerald-600'>{tracker.GoalQuantity}</span> application{tracker.GoalQuantity > 1 && 's'} for a lootbox
    </div>
    <div className='space-y-2'>
      <div className='flex justify-between text-xs text-gray-500'>
        <span>{completed} of {gq} completed</span>
        <span>{Math.round((completed/gq) * 100)}%</span>
      </div>
      <ul className='flex items-center gap-1.5'>{blocks}</ul>
    </div>
  </div>
  );
}

function DeadlineBar({date}: {date: Date}) {
  const { tracker, error } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;

  const deadline = formatDate(tracker.CycleDeadline)
  let deadlineMinutes = 60 * 24
  if (tracker.CycleFrequency == 'weekly') {
    deadlineMinutes = 60 * 24 * 7
  }

  let minsRemaining = Math.floor((deadline.valueOf() - date.valueOf())/ 1000 / 60)
  const progressPercentage = Math.max(0, Math.min(100, (minsRemaining / deadlineMinutes) * 100))

  return (<div className='bg-white rounded-xl p-4 shadow-md border border-gray-100 min-w-0 flex-1'>
    <div className='flex items-center justify-between mb-3'>
      <h3 className='text-lg font-semibold text-gray-800 flex items-center gap-2'>
        <span className='w-2 h-2 bg-amber-500 rounded-full'></span>
        Cycle Deadline
      </h3>
      <span className='text-xs font-medium text-gray-500 bg-gray-50 px-2 py-1 rounded-full'>
        {tracker.CycleFrequency}
      </span>
    </div>
    <div className='text-sm text-gray-600 mb-3'>
      Progress resets in <span className='font-semibold text-amber-600'>{formatMinutes(minsRemaining)}</span>
    </div>
    <div className='space-y-2'>
      <div className='flex justify-between text-xs text-gray-500'>
        <span>Progress resets in {formatMinutes(minsRemaining)}</span>
        <span>{Math.round(progressPercentage)}%</span>
      </div>
      <div className='w-full bg-gray-200 rounded-full h-3 overflow-hidden'>
        <div
          className='h-full bg-gradient-to-r from-amber-400 to-amber-500 rounded-full transition-all duration-500 ease-out'
          style={{width: `${progressPercentage}%`}}
        ></div>
      </div>
    </div>
  </div>)
}

function DailyStreak({date}: {date: Date}) {
  const { tracker, serverDay, error } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;

  const maxProg = 25 * 60;

  // progress bar deadline is set to server's "today" + 25h
  const deadline =  serverDay.valueOf() + (1000 * 25)
  let minsRemaining = Math.floor((deadline - date.valueOf())/ 1000 / 60)
  const progressPercentage = tracker.CurDailyStreak > 0 ? Math.max(0, Math.min(100, (minsRemaining / maxProg) * 100)) : 0

  return (<div className='bg-white rounded-xl p-4 shadow-md border border-gray-100 min-w-0 flex-1'>
    <div className='flex items-center justify-between mb-3'>
      <h3 className='text-lg font-semibold text-gray-800 flex items-center gap-2'>
        <span className='w-2 h-2 bg-orange-500 rounded-full'></span>
        Daily Streak
      </h3>
      <div className='flex items-center gap-2'>
        <span className='text-xs font-medium text-gray-500 bg-gray-50 px-2 py-1 rounded-full'>
          {tracker.CurDailyStreak > 0 ? tracker.CurDailyStreak : '0'}
        </span>
        {tracker.IsLitDailyStreak === true
          ? <img src={flameHot} alt="daily streak met" className="h-6 w-6" />
          : <img src={flameCold} alt="daily streak not met" className="h-6 w-6" />
        }
      </div>
    </div>
    
    <div className='space-y-2'>
      <div className='flex justify-between text-xs text-gray-500'>
      <div className='text-sm text-gray-500 mb-3'>
      {tracker.CurDailyStreak > 0
        ? <>
            {tracker.IsLitDailyStreak ? 'Next cycle begins' : 'Streak expires'} in <span className='font-semibold text-orange-600'>{formatMinutes(minsRemaining)}</span>
          </>
        : 'No applications completed today'
      }
    </div>
        <span>{Math.round(progressPercentage)}%</span>
      </div>
      <div className='w-full bg-gray-200 rounded-full h-3 overflow-hidden'>
        <div
          className='h-full bg-gradient-to-r from-orange-400 to-orange-500 rounded-full transition-all duration-500 ease-out'
          style={{width: `${progressPercentage}%`}}
        ></div>
      </div>
    </div>
  </div>)
}

function ProgressTracker( {onSettingsClick}: {onSettingsClick: () => void }) {
  const isDesktop = useScreenSize()
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
  return (
    <div className='w-full'>
      <div className='flex items-center justify-between mb-4'>
        {/* <h2 className='text-xl font-bold text-gray-800'>Progress Dashboard</h2> */}
        <div></div>
        <button
          className='bg-white hover:bg-gray-50 border border-gray-200 rounded-xl p-3 shadow-md transition-all duration-200 hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1'
          onClick={onSettingsClick}
          title="Tracker Settings"
        >
          <img src={gearIcon} alt="Settings" className="h-5 w-5" />
        </button>
      </div>

      <div className={`grid gap-4 ${isDesktop ? 'grid-cols-3' : 'grid-cols-1 sm:grid-cols-2'}`}>
        <ProgressBoxes />
        <DeadlineBar date={date} />
        <div className={isDesktop ? '' : 'sm:col-span-2'}>
          <DailyStreak date={date} />
        </div>
      </div>
    </div>
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
  const onSettingsClick = () => {
    isEditing? exitEditing(): setIsEditing(true)
  }
  
  return (      
      <div className="min-h-[79px] bg-slate-600 border-x-violet-300 border-4 w-full mx-auto px-4">
        <div className='flex justify-between px-1.5 mb-1 text-center'>
          <ProgressTracker onSettingsClick={onSettingsClick}/>
          {/* <p className='text-sm p-1'>Current lootboxes earned {(tracker.CycleFrequency === "weekly")? 'this week' : 'today'}: {tracker.CurBoxesAwarded}</p> */}
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
            className="ml-2 rounded-[2vw] text-sm px-1 py-0.5 border-2 opacity-35 bg-slate-600 hover:opacity-100 hover:bg-rose-400"
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
          className="rounded-4xl select-none mt-2 mb-1 py-1 px-3 border-2 bg-slate-500 justify-self-center text-md hover:bg-slate-600"
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
