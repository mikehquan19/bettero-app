import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  TimeScale,
} from 'chart.js';
import { Line } from 'react-chartjs-2';
import 'chartjs-adapter-date-fns';
import { reformatDate } from '../../utils';

// register the tool to be able to to use it in the hand-shakeable way 
ChartJS.register(
  CategoryScale, LinearScale, PointElement, LineElement, TimeScale, Title, Tooltip, Legend
);

// the chart showing that 
export default function TimeseriesChart ({ 
  contentType = "expense", stockSymbol = null, 
  intervalType = "month", timeseriesObject = null 
}) {

  // labels and data of the timeseries chart 
  const timeseriesLabels = [];
  const timeseriesData = [];

  if (timeseriesObject) {
    Object.keys(timeseriesObject).forEach((date) => {
      timeseriesLabels.push(new Date(reformatDate(date, '/', '-')));
      timeseriesData.push(timeseriesObject[date]);
    });
  }

  let chartTitle = "User's expense this " + intervalType;
  if (contentType === "stock price") {
    chartTitle = `${stockSymbol}'s price (starting last month)`
  }

  // data of the chart 
  const chartData = {
    labels: timeseriesLabels,
    datasets: [
      {
        label: "expense on the given date",
        data: timeseriesData,
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
        type: 'time',
        time: {
          unit: 'day',
          unitStepSize: 1,
          displayFormats: {
            'day': 'MMM dd',
          },
        },
        grid: {
          // change the color of the vertical grid of white
          color: "skyblue",
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