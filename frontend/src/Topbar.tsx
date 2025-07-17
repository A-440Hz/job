import { useState } from 'react';
function Glyph() {
  const [source, setSource] = useState("src/assets/sb1rb.png");
  const handleMouseDown = () => {setSource("src/assets/sb2rb.png")};
  const handleMouseUp = () => {setSource("src/assets/sb1rb.png")};
  return <img className="flex select-none min-w-12 max-w-24 mr-0" id="glyph" onMouseDown={handleMouseDown} onMouseUp={handleMouseUp} src={source}/> 
}

export function Navbar() {
  return <span className='border-2 justify-self-end'>
    <a className='text-sm/6' href='user'>
      some links this should be an a actually
    </a>
    <span>
      some links
    </span>
    <span>
      some links
    </span>
  </span>;
} 

export default function Topbar() {
    return (
    <div className='flex h-[62px] items-center border justify-between select-none'>
      <span> <Glyph /> </span>
      <span className="flex border select-none text-4xl ml-0"> Hi this is a topbar </span>
      <Navbar />
    </div>);
}

