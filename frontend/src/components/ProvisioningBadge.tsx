import React from 'react';
import { RefreshCw, CheckCircle, AlertCircle, Clock, Loader2 } from 'lucide-react';

interface ProvisioningBadgeProps {
  status: 'provisioned' | 'failed' | 'pending' | 'not_provisioned' | string;
  onRetry?: () => void;
}

const ProvisioningBadge: React.FC<ProvisioningBadgeProps> = ({ status, onRetry }) => {
  switch (status) {
    case 'provisioned':
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 border border-green-200">
          <CheckCircle className="w-3 h-3 mr-1" />
          Provisioned
        </span>
      );
    case 'failed':
      return (
        <div className="flex items-center space-x-2">
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 border border-red-200">
            <AlertCircle className="w-3 h-3 mr-1" />
            Failed
          </span>
          {onRetry && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onRetry();
              }}
              title="Retry Provisioning"
              className="p-1 hover:bg-gray-100 rounded-full transition-colors text-gray-500 hover:text-blue-600"
            >
              <RefreshCw className="w-3 h-3" />
            </button>
          )}
        </div>
      );
    case 'pending':
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 border border-yellow-200">
          <Loader2 className="w-3 h-3 mr-1 animate-spin" />
          Pending
        </span>
      );
    case 'not_provisioned':
    default:
      return (
        <div className="flex items-center space-x-2">
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800 border border-gray-200">
            <Clock className="w-3 h-3 mr-1" />
            Not Provisioned
          </span>
           {onRetry && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onRetry();
              }}
              title="Start Provisioning"
              className="p-1 hover:bg-gray-100 rounded-full transition-colors text-gray-500 hover:text-blue-600"
            >
              <RefreshCw className="w-3 h-3" />
            </button>
          )}
        </div>
      );
  }
};

export default ProvisioningBadge;
