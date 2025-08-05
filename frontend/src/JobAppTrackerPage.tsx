import { useState, useEffect } from 'react'
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, updateTrackerItem, formatDate } from './api/tracker'
import { useTrackerData } from './JobAppTrackerDataContext';
import { useScreenSize } from './ScreenSizeProvider';
import './App.css'

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

function ProgressTracker() {
  const { tracker, error, setTracker } = useTrackerData();
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
    <span>
      <p className='text-sm'> Remaining applications until next lootbox: {tracker.GoalQuantity - tracker.CurScorableItems} </p>
      <p className='text-sm'> Time: {date.valueOf()} </p>
      <p className='text-sm m-0 p-0'> Next deadline: {Math.floor((deadline.valueOf() - date.valueOf())/ 1000 / 60)} minutes </p>
      <progress value={minsRemaining} max={deadlineMinutes}></progress>
    </span>
  );
}

function TrackerBar() {

  return (
      
      <div className="text-xl mb-1 min-h-[79px] bg-blue-400 border-x-violet-300 border-4 text-center">
        <ProgressTracker />
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
