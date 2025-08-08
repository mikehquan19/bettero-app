import {
  Chart as ChartJS, CategoryScale, LinearScale, PointElement,
  LineElement, Title, Tooltip, Legend, TimeScale,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import { reformatDate } from '@utils';

// register the tool to be able to to use it in the hand-shakeable way 
ChartJS.register(
  CategoryScale, LinearScale, PointElement, LineElement, TimeScale, Title, Tooltip, Legend
);

interface TimeseriesChartProps {
  contentType: string, 
  intervalType: string, 
  stockSymbol?: string,
  timeSeriesObject?: Record<string, number>
}

// the chart showing that 
export default function TimeseriesChart ({ 
  contentType = "expense", intervalType = "month", stockSymbol, timeSeriesObject
}: TimeseriesChartProps) {

  // labels and data of the timeseries chart 
  const timeSeriesLabels: Date[] = [];
  const timeSeriesData: number[] = [];

  if (timeSeriesObject) {
    Object.keys(timeSeriesObject).forEach((date: string) => {
      timeSeriesLabels.push(new Date(reformatDate(date, '/', '-')));
      timeSeriesData.push(timeSeriesObject[date]);
    });
  }

  let chartTitle = "User's expense this " + intervalType;
  if (contentType === "stock price") {
    chartTitle = `${stockSymbol}'s price (starting last month)`
  }

  // data of the chart 
  const chartData = {
    labels: timeSeriesLabels,
    datasets: [
      {
        label: "expense on the given date",
        data: timeSeriesData,
        borderColor: "rgb(0, 0, 0, 0.5)",
      }
    ],
  };

  // options of the chart
  const chartOptions = {
    // this is the gain control of the width and height of the canvas
    maintainAspectRatio: false,
    responsive: true,
    scales: {
      x: { // so that the horizontal axis will display unit of time
        type: "time" as const,
        time: {
          unit: "day" as const,
          displayFormats: {
            'day': 'MMM dd',
          },
        },
        grid: {
          color: "skyblue", // change the color of the vertical grid of white
        },
      },
    },
    plugins: {
      // set the title for the chart 
      title: {
        display: true,
        color: 'black',
        text: chartTitle,
        font: {
          size: 20,
        }
      }
    },
  };

  return <Line data={chartData} options={chartOptions} />
}