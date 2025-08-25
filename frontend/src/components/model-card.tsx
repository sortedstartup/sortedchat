import { useStore } from "@nanostores/react";
import { useState } from "react";
import { DownloadStatus, Model as ModelType } from '../../proto/inferenceservice';
import { $downloadingModels, downloadModel } from "@/store/inference";


export const ModelCard = ({ model }: { model: ModelType }) => {
  const downloadingModels = useStore($downloadingModels);
  const [showUrl, setShowUrl] = useState(false);

  const isDownloading = downloadingModels.has(model.name);
  const isDownloaded = model.is_downloaded;
  const isDownloadable = model.is_downloadable;

  const handleDownload = async () => {
    if (!isDownloadable || isDownloaded || isDownloading) return;

    try {
      await downloadModel(model.name);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const getProgressData = () => {
    try {
      if (!model.progress) return null;

      return {
        filesize: model.progress.file_size || 0,
        status: model.progress.status || 0,
        progress: model.progress.progress || 0,
        speed: model.progress.speed || 0
      };
    } catch (error) {
      console.error('Error getting progress data:', error);
      return null;
    }
  };

  const progressData = getProgressData();

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(2)} KB`;
    if (bytes < 1024 ** 3) return `${(bytes / 1024 ** 2).toFixed(2)} MB`;
    if (bytes < 1024 ** 4) return `${(bytes / 1024 ** 3).toFixed(2)} GB`;
    return `${(bytes / 1024 ** 4).toFixed(2)} TB`;
  };

  const formatSpeed = (kbps: number): string => {
    if (kbps === 0) return '0 KB/s';
    if (kbps < 1024) {
      return parseFloat(kbps.toFixed(1)) + ' KB/s';
    }
    const mbps = kbps / 1024;
    return parseFloat(mbps.toFixed(1)) + ' MB/s';
  };

  const getButtonState = () => {
    // Check if actively downloading (status 2) or from local state
    const isActivelyDownloading = isDownloading || (progressData?.status === 2);

    if (isActivelyDownloading) {
      return {
        text: 'Downloading...',
        disabled: true,
        className: 'bg-blue-500 text-white cursor-not-allowed opacity-75'
      };
    }

    if (isDownloaded || progressData?.status === DownloadStatus.COMPLETED) {
      return {
        text: 'Downloaded',
        disabled: true,
        className: 'bg-gray-400 text-white cursor-not-allowed opacity-50'
      };
    }

    if (progressData?.status === DownloadStatus.FAILED) {
      return {
        text: 'Failed - Retry',
        disabled: false,
        className: 'bg-red-500 hover:bg-red-600 text-white cursor-pointer'
      };
    }

    if (!isDownloadable) {
      return null;
    }

    return {
      text: 'Download',
      disabled: false,
      className: 'bg-green-500 hover:bg-green-600 text-white cursor-pointer'
    };
  };

  const buttonState = getButtonState();

  return (
    <div className="bg-white rounded-lg shadow-md p-6 border border-gray-200 hover:shadow-lg transition-shadow">
      <div className="flex justify-between items-start mb-4">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">{model.name}</h3>
          <div className="space-y-1 text-sm text-gray-600">
            <p><span className="font-medium">Provider:</span> {model.provider}</p>
            <p><span className="font-medium">ID:</span> {model.id}</p>
            {model.url && (
              <div>
                <button
                  onClick={() => setShowUrl(!showUrl)}
                  className="text-blue-500 hover:text-blue-700 text-sm font-medium"
                >
                  {showUrl ? 'Hide URL' : 'Show URL'}
                </button>
                {showUrl && (
                  <p className="mt-1">
                    <span className="font-medium">URL:</span>
                    <span className="ml-1 text-gray-600 break-all font-mono text-xs">
                      {model.url}
                    </span>
                  </p>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="ml-4 flex flex-col items-end space-y-2">
          {buttonState && (
            <button
              onClick={handleDownload}
              disabled={buttonState.disabled}
              className={`px-4 py-2 rounded-md font-medium transition-colors ${buttonState.className}`}
            >
              {buttonState.text}
            </button>
          )}

          {progressData && progressData.status === 2 && (
            <div className="w-48">
              <div className="flex items-center justify-between text-xs text-gray-600 mb-1">
                <span>{progressData.progress}%</span>
                <span>{formatFileSize(progressData.filesize)}</span>
                {progressData.speed > 0 && (
                  <span>{formatSpeed(progressData.speed)}</span>
                )}
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${progressData.progress}%` }}
                />
              </div>
            </div>
          )}
        </div>
      </div>

      {(model.input_token_cost > 0 || model.output_token_cost > 0) && (
        <div className="border-t pt-4 mt-4">
          <h4 className="text-sm font-medium text-gray-700 mb-2">Token Costs</h4>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-gray-600">Input:</span>
              <span className="ml-2 font-mono">${model.input_token_cost.toFixed(6)}</span>
            </div>
            <div>
              <span className="text-gray-600">Output:</span>
              <span className="ml-2 font-mono">${model.output_token_cost.toFixed(6)}</span>
            </div>
          </div>
        </div>
      )}

      <div className="border-t pt-4 mt-4">
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center space-x-2 flex-wrap">
            {/* Show downloadable/local status for downloadable models */}
            {isDownloadable && (
              <>
                <span className="px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  Local
                </span>
                {isDownloaded && (
                  <span className="px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    Downloaded
                  </span>
                )}
              </>
            )}

            {!isDownloadable && (
              <div className="flex items-center space-x-2">
                <span className="px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                  Remote
                </span>
                {model.status !== 0 && (
                  <span className="text-xs text-gray-600">
                    Model Status: {model.status}
                  </span>
                )}
              </div>
            )}
          </div>

          {progressData && progressData.filesize > 0 && progressData.status !== 2 && (
            <span className="text-gray-500">{formatFileSize(progressData.filesize)}</span>
          )}
        </div>

        {progressData && progressData.status === 4 && (
          <div className="mt-2 text-xs text-red-600">
            Download failed • Click to retry
          </div>
        )}
      </div>
    </div>
  );
};