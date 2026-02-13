import { useEffect, useState, useRef } from 'react';
import { useTrackerData } from './JobAppTrackerDataContext';
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, restoreTrackerItem, updateTrackerItem } from './api/tracker';
import { useScreenSize } from './ScreenSizeProvider';

interface DeletedItem {
  ID: string;
  timestamp: number;
}

/**
 * a parent component containing the list of job application items
 * @component
 * @returns {JSX.Element} a React component containing the list of items
 */
function ItemsList() {
  const isDesktop = useScreenSize();
  const { tracker, error, refreshData, setTracker } = useTrackerData();
  const [isNewItem, setIsNewItem] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [recentlyDeleted, setRecentlyDeleted] = useState<Array<DeletedItem>>([]);

  // Track in-flight restore operations
  const [restoringIds, setRestoringIds] = useState<Set<string>>(new Set());

  // Track last restore time for debouncing
  const lastRestoreTime = useRef<Record<string, number>>({});
  const RESTORE_COOLDOWN = 1000; // 1 second

  if (error) return <div>Error loading backend: {error}</div>;

  // on mount, load and filter recently deleted items from localStorage
  useEffect(() => {
    const recentlyDeletedStr = localStorage.getItem("recentlyDeleted");
    if (recentlyDeletedStr) {
      const parsedDeletedItems = JSON.parse(recentlyDeletedStr);
      // filter out items older than 24hr (86400000ms)
      const filterLimit = 86400000;
      const filteredDeletedItems = parsedDeletedItems.filter((item: DeletedItem) => Date.now() - item.timestamp < filterLimit);
      setRecentlyDeleted(filteredDeletedItems);

      // Update localStorage if items were filtered out
      if (filteredDeletedItems.length !== parsedDeletedItems.length) {
        localStorage.setItem("recentlyDeleted", JSON.stringify(filteredDeletedItems));
      }
    }
  }, []);

  useEffect(() => {
    refreshData()
  }, [tracker?.UserID])

  const blankItem = { Title: "", Body: "", Url: "" };

  const handleNew = async (newTitle: string, newBody: string, newUrl: string) : Promise<boolean> => {
    try {
      const newData = await createTrackerItem(newTitle, newBody, newUrl);
      if (newData.tracker) {
        setTracker(newData.tracker);
        return true;
      } else {
        console.error("Error finding data from Backend");
      }
    } catch (error: any) {
      console.error(error.message);
    }
    return false;
  };

  const handleEdit = async (item: any, newTitle: string, newBody: string, newUrl: string) : Promise<boolean> => {
    if (item.Title === newTitle && item.Body === newBody && item.Url === newUrl) return false;
    try {
      const newData = await updateTrackerItem(item.ID, newBody, newTitle, newUrl);
      if (newData.tracker) {
        setTracker(newData.tracker);
        return true;
      } else {
        console.error("Error finding data from Backend");
      }
    } catch (error: any) {
      console.error(error.message);
    }
    return false;
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteTrackerItem(id);
      const deletedItem: DeletedItem = { ID: id, timestamp: Date.now() };
      const updatedDeleted = [...recentlyDeleted, deletedItem];
      setRecentlyDeleted(updatedDeleted);
      localStorage.setItem("recentlyDeleted", JSON.stringify(updatedDeleted));
    } catch (error: any) {
      console.error(error.message);
    }
    refreshData();
  };

  const handleRestore = async (id: string) => {
    // Guard 1: Prevent double-restore (check if already in progress)
    if (restoringIds.has(id)) {
      console.log('Restore already in progress for item:', id);
      return;
    }

    // Guard 2: Debounce - prevent rapid successive clicks
    const now = Date.now();
    const lastTime = lastRestoreTime.current[id] || 0;
    if (now - lastTime < RESTORE_COOLDOWN) {
      console.log('Please wait before restoring again');
      return;
    }
    lastRestoreTime.current[id] = now;

    // Guard 3: Verify item exists in deleted list
    const deletedItem = recentlyDeleted.find(item => item.ID === id);
    if (!deletedItem) {
      console.error('Item not found in deleted list:', id);
      return;
    }

    // Mark as restoring (disables button)
    setRestoringIds(prev => new Set(prev).add(id));

    // Optimistic update: remove from deleted list immediately
    const updatedDeleted = recentlyDeleted.filter(item => item.ID !== id);
    setRecentlyDeleted(updatedDeleted);
    localStorage.setItem("recentlyDeleted", JSON.stringify(updatedDeleted));

    try {
      const newData = await restoreTrackerItem(id);

      if (newData.tracker) {
        setTracker(newData.tracker);
        // Success - item stays removed from recentlyDeleted
      } else {
        throw new Error('No tracker data returned');
      }
    } catch (error: any) {
      console.error('Restore failed:', error.message);

      // Rollback optimistic update on error - add the item back
      const rolledBackDeleted = [...updatedDeleted, deletedItem];
      setRecentlyDeleted(rolledBackDeleted);
      localStorage.setItem("recentlyDeleted", JSON.stringify(rolledBackDeleted));

      // Show error alert
      alert(`Failed to restore item: ${error.message}`);
    } finally {
      // Always clean up in-flight tracking
      setRestoringIds(prev => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }
  };

  return (
    <div className='justify-self-center border-2 w-full max-w-170'>
      {/* Undo notification bar */}
      {recentlyDeleted.length > 0 && (
        <div className='bg-amber-50 border-l-4 border-amber-400 p-3 mx-2 mt-2 rounded'>
          <div className='flex items-center justify-between gap-4'>
            <span className='text-sm text-amber-800'>
              {recentlyDeleted.length} item{recentlyDeleted.length > 1 ? 's' : ''} deleted
            </span>
            <button
              className='text-sm font-medium text-amber-600 hover:text-amber-800 underline disabled:opacity-50 disabled:cursor-not-allowed transition-opacity'
              onClick={() => handleRestore(recentlyDeleted[recentlyDeleted.length - 1].ID)}
              disabled={restoringIds.has(recentlyDeleted[recentlyDeleted.length - 1].ID)}
              title='Restore most recently deleted item'
            >
              {restoringIds.has(recentlyDeleted[recentlyDeleted.length - 1].ID)
                ? 'Restoring...'
                : 'Undo'}
            </button>
          </div>
        </div>
      )}

      {isNewItem ? (
        <Item
          item={blankItem}
          isEditing={editingId === "new"}
          isNewItem={true}
          setIsNewItem={setIsNewItem}
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
            setEditingId={setEditingId}
            handleEdit={handleEdit}
            handleDelete={handleDelete}
            handleNew={handleNew}
          />
        ))}
      </ul>
    </div>
  );
}

export default ItemsList;
