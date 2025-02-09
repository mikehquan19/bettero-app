import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  ArcElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { useRef } from 'react';
import { Doughnut } from 'react-chartjs-2';

// register the tool to be able to to use it in the hand-shakeable way 
ChartJS.register(CategoryScale, LinearScale, ArcElement, Title, Tooltip, Legend);

// chart indicating the composition percentage of this month  
const CompositionChart = ({
  intervalType = "month",
  compositionObject,
  handleCategoryClick = null,
  chartType = "regular" }) => {

  const chartRef = useRef();

  // list of categories of expense for labels 
  const categories = Object.keys(compositionObject);
  const compositionData = Object.values(compositionObject);

  // data of the chart 
  const chartData = {
    labels: categories,
    datasets: [
      {
        label: "composition of this month",
        data: compositionData,
        backgroundColor: [
          'rgb(255, 99, 132)',
          'rgb(54, 162, 235)',
          'rgb(255, 206, 86)',
          'rgb(75, 192, 192)',
          'rgb(153, 102, 255)',
          'rgb(255, 159, 64)',
          'rgb(126, 65, 255)',
        ],
        hoverBackgroundColor: "rgb(0, 0, 0, 0.5)",
        hoverBorderWidth: 2,
        hoverBorderColor: "white",
      }
    ],
  };

  // determine the title of the chart based on the type of the chart
  let chartTitle = "";
  let titleSize = 0;
  if (chartType === "regular") {
    titleSize = 20;
    chartTitle = 'Composition of categories of this ' + intervalType;
  } else if (chartType === "budget") {
    titleSize = 18;
    chartTitle = ['Budget percentage of each category', 'on the total budget'];
  } else {
    titleSize = 18;
    chartTitle = ['Current percentage of each category', 'on the total expense'];
  }

  // options of the chart 
  const chartOptions = {
    // this is the gain control of the width and height of the canvas
    maintainAspectRatio: false,
    responsive: true,
    elements: {
      // get rid of the white border of the pie
      arc: {
        borderWidth: 0,
      }
    },
    plugins: {
      // set the title for the chart 
      title: {
        display: true,
        color: "black",
        text: chartTitle,
        font: {
          size: titleSize,
        }
      }
    },
  };

  // render the bar chart 
  return <Doughnut
    onClick={(e) => {
      e.preventDefault();
      handleCategoryClick(e, chartRef);
    }} ref={chartRef} data={chartData} options={chartOptions} />;
}

export default CompositionChart
