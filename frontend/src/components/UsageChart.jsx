import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';

const UsageChart = ({ data, type = 'line', title, description, dataKey = 'value', xAxisKey = 'date' }) => {
  const formatYAxis = (value) => {
    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`;
    }
    if (value >= 1000) {
      return `${(value / 1000).toFixed(1)}K`;
    }
    return value.toString();
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  return (
    <Card>
      <CardHeader>
        {title && <CardTitle>{title}</CardTitle>}
        {description && <CardDescription>{description}</CardDescription>}
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          {type === 'line' ? (
            <LineChart data={data}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis 
                dataKey={xAxisKey} 
                tickFormatter={formatDate}
                className="text-xs"
              />
              <YAxis 
                tickFormatter={formatYAxis}
                className="text-xs"
              />
              <Tooltip 
                labelFormatter={formatDate}
                formatter={(value) => [formatYAxis(value), dataKey]}
              />
              <Legend />
              <Line 
                type="monotone" 
                dataKey={dataKey} 
                stroke="#0066FF" 
                strokeWidth={2}
                dot={{ fill: '#0066FF', r: 4 }}
                activeDot={{ r: 6 }}
              />
            </LineChart>
          ) : (
            <BarChart data={data}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis 
                dataKey={xAxisKey} 
                tickFormatter={formatDate}
                className="text-xs"
              />
              <YAxis 
                tickFormatter={formatYAxis}
                className="text-xs"
              />
              <Tooltip 
                labelFormatter={formatDate}
                formatter={(value) => [formatYAxis(value), dataKey]}
              />
              <Legend />
              <Bar 
                dataKey={dataKey} 
                fill="#0066FF"
                radius={[8, 8, 0, 0]}
              />
            </BarChart>
          )}
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
};

export default UsageChart;
