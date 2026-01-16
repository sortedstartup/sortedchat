import { DownloadStatus, Model as ModelType } from '../../proto/inferenceservice';
import { downloadModel, cancelDownload, deleteModel } from "@/store/inference";
import { Download, X, Trash2, Bot, Box, ArrowUp, ArrowDown } from "lucide-react";

export const ModelCard = ({ model, isLocal = false }: { model: ModelType; isLocal?: boolean }) => {
  const isDownloaded = model.is_downloaded;
  // For local models, treat as downloadable even if flag is not set
  const isDownloadable = isLocal ? true : model.is_downloadable;

  const handleDownload = async () => {
    if (!isDownloadable || isDownloaded || model.status === DownloadStatus.DOWNLOADING) return;

    try {
      await downloadModel(model.name);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const handleCancel = async () => {
    try {
      await cancelDownload(model.name);
    } catch (error) {
      console.error('Cancel failed:', error);
    }
  };

  const handleDelete = async () => {
    try {
      await deleteModel(model.name);
    } catch (error) {
      console.error('Delete failed:', error);
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
    if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1024 ** 3) return `${(bytes / 1024 ** 2).toFixed(1)} MB`;
    if (bytes < 1024 ** 4) return `${(bytes / 1024 ** 3).toFixed(1)} GB`;
    return `${(bytes / 1024 ** 4).toFixed(1)} TB`;
  };

  const isDownloading = progressData?.status === DownloadStatus.DOWNLOADING;
  const isEmbedding = model.is_embedding_model;
  const hasModelInfo = model.provider === 'local' && model.model_info;

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-4 flex items-center justify-between hover:shadow-md transition-shadow">
      {/* Left Side */}
      <div className="flex items-center gap-4">
        <div className="w-12 h-12 rounded-xl bg-muted/50 flex items-center justify-center border border-border/50">
          {isEmbedding ? (
            <Box className="w-6 h-6 text-foreground/70" />
          ) : (
            <Bot className="w-6 h-6 text-foreground/70" />
          )}
        </div>

        <div className="flex flex-col justify-center">
          {isEmbedding && (
            <span className="text-[10px] font-bold text-purple-600 dark:text-purple-400 bg-purple-100 dark:bg-purple-900/30 px-1.5 py-0.5 rounded w-fit mb-0.5 uppercase tracking-wider">
              Embedding
            </span>
          )}
          <h3 className="font-bold text-base text-foreground leading-tight">{model.name}</h3>

          {/* Provider and Quantization */}
          <div className="flex items-center gap-2 mt-0.5">
            {hasModelInfo && model.model_info?.quantization && (
              <span className="text-[10px] font-mono font-bold text-blue-600 dark:text-blue-400 bg-blue-100 dark:bg-blue-900/30 px-1.5 py-0.5 rounded">
                {model.model_info.quantization}
              </span>
            )}
            {hasModelInfo && model.model_info?.download_size && (
              <span className="text-[10px] text-muted-foreground font-medium">{model.model_info.download_size}</span>
            )}
          </div>

          {/* Creator and Modified By */}
          {(model.creator_name || model.modified_by) && (
            <div className="flex items-center gap-1.5 mt-0.5 text-[10px] text-muted-foreground">
              {model.creator_name && model.provider === 'local' && (
                <span className="font-medium">{model.creator_name}</span>
              )}
              {model.creator_name && model.modified_by && (
                <span>•</span>
              )}
              {model.modified_by && (
                <span className="italic">modified by {model.modified_by}</span>
              )}
            </div>
          )}

          {/* Description */}
          {model.description && (
            <p className="text-[10px] text-muted-foreground mt-1 line-clamp-2 max-w-md">
              {model.description}
            </p>
          )}
        </div>
      </div>

      {/* Right Side */}
      <div className="flex items-center gap-4">
        {/* Stats / Progress */}
        {isDownloading ? (
          <div className="flex flex-col items-end min-w-[120px]">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-bold text-purple-600 dark:text-purple-400">
                FETCHING
              </span>
              <span className="text-xs font-bold text-foreground">
                {progressData?.progress}%
              </span>
            </div>
            <div className="w-32 h-1.5 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-purple-600 dark:bg-purple-400 transition-all duration-300 ease-out"
                style={{ width: `${progressData?.progress}%` }}
              />
            </div>
            <div className="flex items-center gap-1 mt-1 text-[10px] text-muted-foreground font-mono">
              <span>{formatFileSize(progressData?.filesize || 0)}</span>
              {progressData?.speed ? (
                <span>/ {(progressData.speed / 1024).toFixed(1)} MB/s</span>
              ) : null}
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-end mr-2">
            {/* {(model.input_token_cost > 0 || model.output_token_cost > 0) && ( */}
            <div className="flex flex-col items-end gap-0.5">
              <div className="flex items-center gap-1.5" title="Input Token Cost">
                <ArrowUp className="w-3 h-3 text-muted-foreground/70" />
                <span className="text-xs font-mono font-medium text-foreground">
                  ${model.input_token_cost}
                </span>
              </div>
              <div className="flex items-center gap-1.5" title="Output Token Cost">
                <ArrowDown className="w-3 h-3 text-muted-foreground/70" />
                <span className="text-xs font-mono font-medium text-foreground">
                  ${model.output_token_cost}
                </span>
              </div>
            </div>
            {/* )} */}
          </div>
        )}

        {/* Action Button */}
        {isDownloadable && (
          <button
            onClick={
              isDownloading ? handleCancel :
                isDownloaded ? handleDelete :
                  handleDownload
            }
            className={`
              w-10 h-10 rounded-full flex items-center justify-center transition-all duration-200
              ${isDownloading
                ? 'bg-purple-600 hover:bg-purple-700 text-white shadow-sm hover:shadow-purple-500/25'
                : isDownloaded
                  ? 'bg-muted hover:bg-destructive hover:text-destructive-foreground text-muted-foreground'
                  : 'bg-primary hover:bg-primary/90 text-primary-foreground shadow-sm hover:shadow-primary/25'
              }
            `}
            title={
              isDownloading ? "Cancel Download" :
                isDownloaded ? "Delete Model" :
                  "Download Model"
            }
          >
            {isDownloading ? (
              <X className="w-5 h-5" />
            ) : isDownloaded ? (
              <Trash2 className="w-5 h-5" />
            ) : (
              <Download className="w-5 h-5" />
            )}
          </button>
        )}
      </div>
    </div>
  );
};