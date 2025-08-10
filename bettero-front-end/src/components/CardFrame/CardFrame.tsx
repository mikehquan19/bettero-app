import { Link } from 'react-router-dom';
import './CardFrame.scss';
import { MouseEvent, ReactNode } from 'react';
import { capitalize } from '@utils';

interface CardFrameProps {
  children: ReactNode,
  cardType: string, 
  numCards: number
}

// CardFrame component , wrapping around Card
export default function CardFrame({ children, cardType, numCards }: CardFrameProps) {
  return (
    <div className="card-frame">
      <h3>
        <Link to={`/accounts/${cardType}`}
          style={{ color: "black", }}
          onMouseOver={(e: MouseEvent<HTMLAnchorElement>) => {
            e.currentTarget.style.color = "white";
          }}
          onMouseOut={(e: MouseEvent<HTMLAnchorElement>) => {
            e.currentTarget.style.color = "black";
          }}
        >
          {capitalize(cardType)} Accounts ({numCards})
        </Link>
      </h3>
      <div className="card-list">{ children }</div>
    </div>
  );
}