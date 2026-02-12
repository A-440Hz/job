import { useEffect, useState } from 'react';
import { useTrackerData } from './JobAppTrackerDataContext';
import Item from './Item';
import { createTrackerItem, deleteTrackerItem, updateTrackerItem } from './api/tracker';
import { useScreenSize } from './ScreenSizeProvider';
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

  if (error) return <div>Error loading backend: {error}</div>;

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
