import { MouseEvent } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCirclePlus } from '@fortawesome/free-solid-svg-icons';
import './Button.scss';

interface ButtonProps {
  content: string, 
  onClick?: (e: MouseEvent<HTMLButtonElement>) => void
}

// the Button
export default function Button({ content, onClick }: ButtonProps) {

  function handleMouseOver(e: MouseEvent<HTMLButtonElement>): void {
    e.currentTarget.style.backgroundColor = "white";
    e.currentTarget.style.border = "2px solid darkcyan";
    e.currentTarget.style.color = "darkcyan";
  }

  function handleMouseOut(e: MouseEvent<HTMLButtonElement>): void {
    e.currentTarget.style.backgroundColor = "darkcyan";
    e.currentTarget.style.border = "1px solid green";
    e.currentTarget.style.color = "white";
  }
  
  return (
    <div style={{ textAlign: "center" }}>
      <button className="add-button"
        onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}
        onClick={onClick}
      >
        <FontAwesomeIcon icon={faCirclePlus} style={{
          marginRight: "5px",
          display: "block",
          margin: "0 auto 5px auto",
        }} /> 
        {content} 
      </button>
    </div>
  );
}