import { Link } from 'react-router-dom';
import './CardFrame.css';

// CardFrame component , wrapping around Card
export default function CardFrame({ children, cardType, numCards }) {
  return (
    <div className="card-frame">
      <h3>
        <Link to={`/accounts/${cardType}`}
          style={{ color: "black", }}
          onMouseOver={(e) => { e.target.style.color = "white" }}
          onMouseOut={(e) => { e.target.style.color = "black" }}>
          {cardType.charAt(0).toUpperCase() + cardType.slice(1)} Accounts ({numCards})
        </Link>
      </h3>
      <div className="card-list">{ children }</div>
    </div>
  );
}