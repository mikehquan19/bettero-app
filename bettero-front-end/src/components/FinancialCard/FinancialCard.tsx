import './FinancialCard.css';
import { FinancialInfo } from '@interface';

interface FinancialCardProps {
  financialData?: FinancialInfo
}

export default function FinancialCard({ financialData }: FinancialCardProps) {
  return (
    <>
      <h2 className="financial-info-title">This month's financial information</h2>
      <div className="financial-cards-wrapper">
        <div className="financial-card">
          <h4>Total balance:</h4>
          <div>${financialData?.totalBalance || 0}</div>
        </div>
        <div className="financial-card">
          <h4>Amount due:</h4>
          <div>${financialData?.totalAmountDue || 0}</div>
        </div>
        <div className="financial-card">
          <h4>Total income:</h4>
          <div>${financialData?.totalIncome || 0}</div>
        </div>
        <div className="financial-card">
          <h4>Total expense:</h4>
          <div>${financialData?.totalExpense || 0}</div>
        </div>
      </div>
    </>
  );
}