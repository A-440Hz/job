import { useState, useEffect } from 'react'
import { NavLink } from 'react-router-dom';
import ItemsList from './ItemsList';
import LoadingSpin from './LoadingSpin';
import { updateTracker } from './api/tracker'
import { formatDate, dateToInputString, inputStringToDate, adjustTimezoneOffset, convertToBackendTime } from './api/datetime'
import { useTrackerData } from './JobAppTrackerDataContext';
import { useScreenSize } from './ScreenSizeProvider';
import { wakeupScraper } from './api/scraper';
import flameHot from '/flame-hot-svgrepo-com.svg';
import flameCold from '/flame-cold-svgrepo-com.svg';
import gearIcon from '/gear-svgrepo-com.svg';

import './App.css'

/** 
 * converts minutes to a string in the format "Xd Yh Zm"
 * @param {number} minLeft - number of minutes
 * @returns {string} formatted string
 */
function formatMinutes(minLeft:number) {
  if (minLeft < 0) {return -1}
  let d = Math.floor(minLeft/60/24)
  let h = Math.floor(minLeft/60 - d * 24);
  let m = (minLeft % 60);
  // console.log(d, h, m)
  return ((d > 0)? d.toString() + "d ": "") + ((h > 0)? h.toString() + "h ": "") + ((m > 0)? m.toString() + "m ": "");
}

/**
 * Job Application Tracker Page.
 * Displays a list of job applications and their statuses.
 *
 * @component
 * @returns {JSX.Element} The rendered component.
 */
function JobAppTrackerPage() {
  const { user, tracker, error, refreshData } = useTrackerData();
  const isDesktop = useScreenSize();

  // refresh data when user changes (login/logout),
  // trying to avoid refreshData causing a fetch loop.
  useEffect(() => {
    refreshData();
    // send health ping to microservice so it is ready to run
    wakeupScraper()
      .then((data) => {
        if (data && data.status) {
          console.log("health status from microservice: ", data.status);
        }
      })
      .catch((err) => {
        console.log("error from microservice: ", err);
      });
  }, [user?.ID]);

  if (error) return <div>Error loading backend: {error}</div>;
  if (!user || !tracker) return (<>
    <div className='text-center justify-self-center'>I'm on the free version</div>
    <div className="absolute z-100 left-1/2 top-1/2"> <LoadingSpin /> </div>
  </>);

  return (
    <div className="grid mx-auto px-8 pt-4 min-w-8/10 max-w-10/10 justify-self-center mt-3">
      <h1 className="vp-mid mb-4 text-5xl font-semibold select-none text-center text-blue-200 text-shadow-2xs text-shadow-violet-900">
        {isDesktop ?
          (<>
            Job App Tracker With
            <div className="text-5xl font-bold tdrop-shadow-md">Lootboxes</div>
            <div className="text-4xl font-semibold">and Agentic Functionality</div>
          </>)
            :
          'Job App Tracker'}
      </h1>
      <div className={`${isDesktop?'':'w-[90vw] relative left-1/2 right-1/2 -mx-[45vw]'} `}>
        <TrackerBar />
      </div>
      <ItemsList />
    </div>
  );
}

function ProgressBoxes() {
  const { tracker, user, error, refreshData } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;

  const gq = tracker.GoalQuantity;
  const completed = tracker.CurScorableItems % gq;
  const blocks = [];
  const [hasLootboxes, setHasLootboxes] = useState(user.inventory?.NumLootboxes > 0);

  useEffect(() => {
    refreshData();
    setHasLootboxes(user.inventory?.NumLootboxes > 0);
  }, [completed, user.inventory?.NumLootboxes, tracker.CurBoxesAwarded])
  
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
      {hasLootboxes ? 
        <NavLink 
          className='text-xs font-medium text-orange-400 bg-gray-50 px-2 py-1 rounded-full transition hover:scale-95 hover:cursor-pointer' 
          to="/Lootbox"> {user.inventory.NumLootboxes} lootboxes available to open </NavLink> :
        <span className='text-xs font-medium text-gray-500 bg-gray-50 px-2 py-1 rounded-full'>
          {tracker.CurBoxesAwarded} earned this deadline
        </span>
      }
    </div>
    <div className='text-sm text-gray-600 mb-3'>
      Complete <span className='font-semibold text-emerald-600'>{tracker.GoalQuantity}</span> application{tracker.GoalQuantity > 1 && 's'} for a lootbox
    </div>
    <div className='space-y-2'>
      {/* <div className='flex justify-between text-xs text-gray-500'>
        <span>{completed} of {gq} completed</span>
        <span>{Math.round((completed/gq) * 100)}%</span>
      </div> */}
      <ul className='flex items-center gap-1.5'>{blocks}</ul>
    </div>
  </div>
  );
}

/**
 * Returns a component detailing the amount of time left before the deadline resets.
 * @component
 * @param {Date} date - the current date from new Date()
 * @returns {JSX.Element} A React component displaying the deadline bar.
 */ 
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
    <div className='space-y-2'>
      <div className='flex justify-between text-xs text-gray-500'>
        <span>Progress resets in <span className='font-semibold text-amber-600'>{formatMinutes(minsRemaining)}</span></span>
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

/**
 * Returns a component detailing the amount of time left before the daily streak resets.
 * @component
 * @param {Date} date - the current date from new Date()
 * @returns {JSX.Element} A React component displaying the daily streak bar.
 */
function DailyStreak({date}: {date: Date}) {
  const { tracker, serverDay, error } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;

  const maxProg = 24 * 60;

  // progress bar deadline is set to server's "today" + 24h
  const deadline =  serverDay.valueOf() + (1000 * 60 * 60 * 24)
  let minsRemaining = Math.floor((deadline - date.valueOf())/ 1000 / 60)
  const progressPercentage = tracker.CurDailyStreak > 0 ? Math.max(0, Math.min(100, (minsRemaining / maxProg) * 100)) : 0

  return (<div className='bg-white rounded-xl p-4 shadow-md border border-gray-100 min-w-0 flex-1'>
    <div className='flex items-center justify-between mb-3'>
      <h3 className='text-lg font-semibold text-gray-800 flex items-center gap-2'>
        <span className='w-2 h-2 bg-orange-500 rounded-full'></span>
        Daily Streak
      </h3>
      <div className='flex items-center gap-2'>
        <span className='text-xs font-medium text-gray-500 bg-gray-50 pl-2 py-1 rounded-full'>
          {tracker.CurDailyStreak > 0 ? tracker.CurDailyStreak : ''}
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
            {tracker.IsLitDailyStreak ? 'Next cycle begins in' : 'Streak expires in'}  <span className='font-semibold text-orange-600'>{formatMinutes(minsRemaining)}</span>
          </>
        : 'No applications completed today'
      }
    </div>
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

/**
 * parent component containing all the dynamic tracker graphics and visual components
 * @component
 * @param {function} onSettingsClick - function to call when settings button is clicked 
 * @returns {JSX.Element} A React component displaying the progress tracker.
 */
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
        <div></div>
        <button
          className='bg-white hover:bg-gray-50 border border-gray-200 rounded-xl p-3 shadow-md transition-all duration-200 hover:shadow-lg hover:cursor-pointer focus:outline-none focus:ring-2 focus:ring-yellow-500 focus:ring-offset-1'
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

/**
 * a parent component contianing the ProgressTracker and the tracker edit menu 
 * @component
 * @returns {JSX.Element} a React component containing the TrackerBar
 */
function TrackerBar() {
  const { tracker, user, error, setTracker, modelName, setModelName, refreshData } = useTrackerData();
  if (error) return <div>Error loading backend: {error}</div>;
  
  const [isEditing, setIsEditing] = useState(false);

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
  const [selectedModel, setSelectedModel] = useState(modelName);

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
    refreshData();
    exitEditing();
  }

  const handleEdit = async (penalty:boolean, frequency?:string, deadline?:string, quantity?:number) => {
    const changes: {
      penalty?: boolean;
      frequency?: string;
      deadline?: number;
      quantity?: number;
    } = {};

    // also handle the AI Model selection
    if (selectedModel !== modelName) {
      setModelName(selectedModel)
    }

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
      {isEditing && (
        <div className='bg-white rounded-xl mt-4 p-6 shadow-lg border border-gray-200'>
          <h3 className='text-lg font-semibold text-gray-800 mb-4'>Tracker Settings</h3>

          <div className='grid grid-cols-1 md:grid-cols-2 gap-6'>
            <div className={`space-y-2 p-4 rounded-lg border-2 transition-colors ${
              quantity !== tracker.GoalQuantity ? 'border-amber-400 bg-amber-50' : 'border-gray-200 bg-gray-50'
            }`}>
              <label className='block text-sm font-medium text-gray-700'>
                Lootbox Goal
              </label>
              <p className='text-xs text-gray-500 mb-2'>Applications needed to earn a lootbox</p>
              <input
                className='w-20 text-emerald-600 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500'
                type="number"
                value={quantity}
                min={1}
                onChange={(e) => setQuantity(e.target.valueAsNumber)}
              />
            </div>

            <div className={`space-y-2 p-4 rounded-lg border-2 transition-colors ${
              deadline !== localizedDeadline ? 'border-amber-400 bg-amber-50' : 'border-gray-200 bg-gray-50'
            }`}>
              <label htmlFor="select-date" className='block text-sm font-medium text-gray-700'>
                Goal Deadline
              </label>
              <p className='text-xs text-gray-500 mb-2'>Deadline to meet Lootbox Goal before it resets</p>
              <input
                className='w-full text-amber-600  px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-amber-500 focus:border-amber-500'
                id='select-date'
                type='datetime-local'
                value={deadline}
                min={dateToInputString(now)}
                onChange={(e) => setDeadline(e.target.value)}
              />
            </div>

            <div className={`space-y-2 p-4 rounded-lg border-2 transition-colors ${
              frequency !== tracker.CycleFrequency ? 'border-amber-400 bg-amber-50' : 'border-gray-200 bg-gray-50'
            }`}>
              <label className='block text-sm font-medium text-gray-700'>
                Reset Frequency
              </label>
              <p className='text-xs text-gray-500 mb-2'>How often the deadline resets</p>
              <select
                className='w-full text-amber-600 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-amber-500 focus:border-amber-500'
                value={frequency}
                onChange={(e) => setFrequency(e.target.value)}
              >
                <option value="daily">Daily</option>
                <option value="weekly">Weekly</option>
              </select>
            </div>

            {/* <div className={`space-y-2 p-4 rounded-lg border-2 transition-colors ${
              penalty !== tracker.MissedGoalPenalty ? 'border-amber-400 bg-amber-50' : 'border-gray-200 bg-gray-50'
            }`}>
              <label className='block text-sm font-medium text-gray-700'>
                Goal Failure Penalty
              </label>
              <p className='text-xs text-gray-500 mb-2'>Apply penalties for missed goals</p>
              <label className='flex items-center space-x-2 cursor-pointer'>
                <input
                  className='w-4 h-4 text-slate-600 bg-gray-100 border-gray-300 rounded focus:ring-slate-500'
                  type="checkbox"
                  checked={penalty}
                  onChange={() => setPenalty(!penalty)}
                />
                <span className='text-sm text-amber-600'>Enable penalty</span>
              </label>
            </div> */}

            <div className={`space-y-2 p-4 rounded-lg border-2 transition-colors ${
                selectedModel !== modelName ? 'border-amber-400 bg-amber-50' :
              'border-gray-200 bg-gray-50'
              }`}>
                <label className='block text-sm font-medium text-gray-700'>
                  AI Model
                </label>
                <p className='text-xs text-gray-500 mb-2'>Model used for
              summarization</p>
                <select
                  className='w-full text-amber-600 px-3 py-2 border border-gray-300
              rounded-md focus:ring-2 focus:ring-amber-500 focus:border-amber-500'
                  value={selectedModel}
                  onChange={(e) => setSelectedModel(e.target.value)}
                >
                  <option value="tngtech/deepseek-r1t2-chimera:free">Deepseek R1T2 Chimera</option>
                  <option value="nvidia/nemotron-nano-12b-v2-vl:free">NVIDIA Nemotron Nano 2 VL</option>
                  <option value="kwaipilot/kat-coder-pro:free">KwaiKAT KAT-Coder-Pro V1</option>
                  <option value="qwen/qwen3-coder:free">Qwen Qwen3-Coder-480B-A35B</option>
                  <option value="openai/gpt-oss-20b:free">OpenAI gpt-oss-20b</option>
                </select>
              </div>
          </div>

          <div className='flex justify-end space-x-3 mt-6 pt-4 border-t border-gray-200'>
            <button
              className='px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-500 transition-colors'
              onClick={(e) => {
                e.stopPropagation();
                setIsEditing(false);
              }}
            >
              Cancel
            </button>
            <button
              className='px-4 py-2 text-sm font-medium text-white bg-emerald-600 border border-transparent rounded-md hover:bg-emerald-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-emerald-500 transition-colors'
              onClick={(e) => {
                e.stopPropagation();
                submitChanges();
              }}
            >
              Save Changes
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default JobAppTrackerPage
