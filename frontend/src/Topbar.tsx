import { useEffect, useRef, useState } from 'react';
import { useCollectablesData } from './CollectablesDataContext';
import { useTrackerData } from './JobAppTrackerDataContext';
import { logoutUser } from './api/user';
import { NavLink } from 'react-router-dom';
import { useLocation, useNavigate } from 'react-router-dom';
import { useScreenSize } from './ScreenSizeProvider';

/**
 * A React component for displaying the site glyph/logo.
 * Changes image on hover and navigates to home on click.
 * @component
 * @returns {JSX.Element} The glyph/logo component.
 */
function Glyph() {
  const [source, setSource] = useState("/sb1rb.png");
  const handleMouseAway: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/sb2rb.png");
  };
  const handleMouseOver: React.MouseEventHandler<HTMLImageElement> = (event) => {
    event.preventDefault();
    setSource("/sb1rb.png");
  };
  const handleMouseClick = () => {
    window.location.href = "/";
  }
  return <img className="flex select-none min-w-12 max-w-24 mr-0 hover:cursor-pointer" id="glyph" 
    onMouseEnter={handleMouseAway} 
    onMouseLeave={handleMouseOver}
    onClick={handleMouseClick} 
    src={source}
    draggable="false"
  /> 
}

/**
 * returns a horizontal navigation bar component or a hamburger menu for mobile
 * @component
 * @param {JSON} user - json representation of the user object
 * @param {function} onLogout - function to call on user logout
 * @returns {JSX.Element} A React component for displaying the navigation bar
 */
export function Navbar({ user, onLogout }: { user: any, onLogout?: () => void }) {
  const isDesktop = useScreenSize();
  const [open, setOpen] = useState(false);

  // use ref to get dropdown menu to behave more naturally
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      const target = event.target as Node;
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(target) &&
        buttonRef.current &&
        !buttonRef.current.contains(target)
      ) {
        setOpen(false);
      }
    }
    function handleEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setOpen(false);
      }
    }
    document.addEventListener("click", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("click", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, []);

  const navClass = "flex items-center text-m/6 hover:opacity-60 ";

  function NavItems({ onClick }: { onClick?: () => void }) {
    return (
      <>
        <NavLink
          to="/"
          className={({ isActive }) =>
            isActive ? navClass + "text-amber-300" : navClass
          }
          onClick={onClick}
        >
          Tracker
        </NavLink>

        <NavLink
          to="/Collection"
          className={({ isActive }) =>
            isActive ? navClass + "text-amber-300" : navClass
          }
          onClick={onClick}
        >
          Collection
        </NavLink>

        <NavLink
          to="/Lootbox"
          className={({ isActive }) =>
            isActive ? navClass + "text-amber-300" : navClass
          }
          onClick={onClick}
        >
          Lootbox
        </NavLink>

        <NavLink
          to="/About"
          className={({ isActive }) =>
            isActive ? navClass + "text-amber-300" : navClass
          }
          onClick={onClick}
        >
          About
        </NavLink>

        <NavLink
          to="/Profile"
          className={({ isActive }) =>
            isActive ? navClass + "text-amber-300" : navClass
          }
          onClick={onClick}
        >
          Profile
        </NavLink>
      </>
    );
  }

  return (
    <nav className={`absolute right-0 border-2 px-3 py-2 rounded-lg bg-slate-700 border-violet-200 flex ${isDesktop && 'left-[100px]'}`}>
      
      {/* DESKTOP NAV */}
      {isDesktop && (
        <div className="flex flex-grow justify-between pl-10">
          <NavItems />
          <Login user={user} onLogout={onLogout} />
        </div>
      )}

      {/* MOBILE HAMBURGER */}
      {!isDesktop && (
        <>
        <button
          ref={buttonRef}
          onClick={() => {setOpen(!open)}}
          className="flex flex-col justify-center items-center space-y-1 pr-2"
        >
          <div className={`h-1 w-6 bg-violet-200 transition-all ${open ? "rotate-45 translate-y-2" : ""}`}></div>
          <div className={`h-1 w-6 bg-violet-200 transition-all ${open ? "opacity-0" : ""}`}></div>
          <div className={`h-1 w-6 bg-violet-200 transition-all ${open ? "-rotate-45 -translate-y-2" : ""}`}></div>
        </button>
        {open && (
        <div ref={dropdownRef} className="absolute top-full left-0 w-full bg-slate-700 border-x-2 border-b-2 border-violet-200 rounded-b-lg flex flex-col p-3 space-y-2 z-40">
          <NavItems onClick={() => setOpen(false)} />
        </div>
      )}
        <Login user={user} onLogout={onLogout} />
        </>
      )}

      
      {/* MOBILE DROPDOWN */}
      
    </nav>
  );
}

/**
 * @component
 * @param {JSON} user - json representation of the user object
 * @param {function} onLogout - function to call on user logout
 * @returns {JSX.Element} a login/register button or logout button
 */
export function Login({ user, onLogout }: { user: any; onLogout?: () => void }) {
  if (!user) return <div></div>;
  const loc = useLocation();
  if (user.Registered === true) {
    return (
      <button
        className="hover:opacity-60"
        id="navbar_sign_in_button"
        onClick={() => onLogout && onLogout()}
      >
        Logout
      </button>
    );
  } else {
    return <NavLink className={`pr-1.5 md:pr-3 hover:opacity-60 ${loc.pathname === "/Login" && 'text-amber-300'}`} id="navbar_sign_in_button"  to="/Login">Login/Register</NavLink>
  }
}

/**
 * the Topbar component that is present across all major views
 * @component
 * @returns {JSX.Element} A React component for displaying the website topbar
 */
export default function Topbar() {
  // const isDesktop = useScreenSize();
  const {user, error, refreshData} = useCollectablesData();
  const {refreshData: refreshTrackerData} = useTrackerData();
  const location = useLocation();
  const navigate = useNavigate();

  const onLogout = async () => {
    try {
      await logoutUser();
    } catch (err) {
      console.error('Logout failed:', err);
    } finally {
      // Clear user preferences from localStorage
      localStorage.removeItem('selectedModel');
      localStorage.removeItem('deletedItems');

      // refresh client state and navigate home
      try {
        await Promise.all([refreshData(), refreshTrackerData()]);
      } catch (e) {
        console.error('refreshData failed after logout', e);
      }
      navigate('/');
    }
  };

  // Refresh on mount and when the route path changes. This keeps the
  // navbar counters (e.g. lootbox count) reasonably up-to-date when the
  // user navigates between pages.
  useEffect(() => {
    refreshData();
    // only re-run when pathname changes
  }, [user?.ID, location.pathname]);

  if (error) return <div>Error loading backend: {error}</div>;
  if (!user) return <div className='text-center justify-self-center pt-5'>Please bear with the loading time...</div>; 
    return (
    <div className='flex h-[62px] relative min-w-full items-center select-none z-50'>
      <Glyph />
      {/* <span className="md:flex select-none font-semibold text-3xl ml-0"> {isDesktop && "haotianswebsite.com"} </span> */}
      <Navbar user={user} onLogout={onLogout}/>
      {/* <Login user={user} onLogout={onLogout} /> */}
    </div>);
}

