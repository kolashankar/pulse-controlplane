import React from 'react';

const Logo = ({ className = '', size = 'md' }) => {
  const sizes = {
    sm: { width: 24, height: 24, text: 'text-lg' },
    md: { width: 32, height: 32, text: 'text-xl' },
    lg: { width: 40, height: 40, text: 'text-2xl' },
    xl: { width: 48, height: 48, text: 'text-3xl' }
  };

  const { width, height, text } = sizes[size] || sizes.md;

  return (
    <div className={`flex items-center gap-2 ${className}`}>
      <svg
        width={width}
        height={height}
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient id="pulseGradient" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#0066FF" />
            <stop offset="100%" stopColor="#00A3FF" />
          </linearGradient>
        </defs>
        {/* Pulse waveform */}
        <path
          d="M2 16 L6 16 L8 8 L10 24 L12 12 L14 20 L16 4 L18 28 L20 16 L22 16 L24 10 L26 22 L28 16 L30 16"
          stroke="url(#pulseGradient)"
          strokeWidth="2.5"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
      <span className={`font-semibold ${text} bg-gradient-to-r from-blue-600 to-blue-400 bg-clip-text text-transparent`}>
        Pulse
      </span>
    </div>
  );
};

export default Logo;
