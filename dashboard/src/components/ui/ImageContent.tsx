import { type FC, useState, useEffect } from 'react';
import { Image as ImageIcon, Download, Maximize2, X, AlertCircle } from 'lucide-react';

interface ImageSource {
  type: 'base64';
  media_type: 'image/jpeg' | 'image/png' | 'image/gif' | 'image/webp';
  data: string;
}

interface ImageContentProps {
  source: ImageSource;
  alt?: string;
  maxHeight?: number;
}

export const ImageContent: FC<ImageContentProps> = ({
  source,
  alt = 'Image content',
  maxHeight = 400,
}) => {
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [loadError, setLoadError] = useState(false);

  // Construct data URI
  const dataUri = `data:${source.media_type};base64,${source.data}`;

  // Get file extension for download
  const extension = source.media_type.split('/')[1] || 'png';

  const handleDownload = () => {
    const link = document.createElement('a');
    link.href = dataUri;
    link.download = `image.${extension}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (loadError) {
    return (
      <div className="flex items-center gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
        <AlertCircle className="w-5 h-5" />
        <div>
          <div className="font-medium">Failed to load image</div>
          <div className="text-sm text-red-600">
            Format: {source.media_type} | Size: {Math.round(source.data.length / 1024)}KB
            base64
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      {/* Main image display */}
      <div className="relative group rounded-lg overflow-hidden border border-gray-200 bg-gray-50">
        {/* Image info badge */}
        <div className="absolute top-2 left-2 z-10 flex items-center gap-1.5 px-2 py-1 bg-black/60 text-white text-xs rounded">
          <ImageIcon className="w-3 h-3" />
          <span>{source.media_type.split('/')[1].toUpperCase()}</span>
        </div>

        {/* Control buttons */}
        <div className="absolute top-2 right-2 z-10 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={() => setLightboxOpen(true)}
            className="p-1.5 bg-black/60 text-white rounded hover:bg-black/80 transition-colors"
            title="View fullscreen"
          >
            <Maximize2 className="w-4 h-4" />
          </button>
          <button
            onClick={handleDownload}
            className="p-1.5 bg-black/60 text-white rounded hover:bg-black/80 transition-colors"
            title="Download image"
          >
            <Download className="w-4 h-4" />
          </button>
        </div>

        {/* The image */}
        <img
          src={dataUri}
          alt={alt}
          onError={() => setLoadError(true)}
          className="w-full object-contain cursor-zoom-in"
          style={{ maxHeight }}
          onClick={() => setLightboxOpen(true)}
        />
      </div>

      {/* Lightbox */}
      {lightboxOpen && (
        <Lightbox
          src={dataUri}
          alt={alt}
          onClose={() => setLightboxOpen(false)}
          onDownload={handleDownload}
        />
      )}
    </>
  );
};

// Lightbox component for fullscreen view
interface LightboxProps {
  src: string;
  alt: string;
  onClose: () => void;
  onDownload: () => void;
}

const Lightbox: FC<LightboxProps> = ({ src, alt, onClose, onDownload }) => {
  // Close on escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/90"
      onClick={onClose}
    >
      {/* Controls */}
      <div className="absolute top-4 right-4 flex gap-2 z-10">
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDownload();
          }}
          className="p-2 bg-white/20 text-white rounded-lg hover:bg-white/30 transition-colors"
          title="Download"
        >
          <Download className="w-5 h-5" />
        </button>
        <button
          onClick={onClose}
          className="p-2 bg-white/20 text-white rounded-lg hover:bg-white/30 transition-colors"
          title="Close"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Image */}
      <img
        src={src}
        alt={alt}
        className="max-w-[90vw] max-h-[90vh] object-contain"
        onClick={(e) => e.stopPropagation()}
      />
    </div>
  );
};
